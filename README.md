# 1. Общая схема

```mermaid
flowchart TD
%% Определение узлов верхнего уровня
    Client[Client]
    Gateway[API Gateway<br><font size=2>auth, rate limit, REST→gRPC]

%% Определение микросервисов
S_User[user]
S_Catalog[catalog]
S_Order[order]
S_Inventory[inventory]
S_Payment[payment]
S_Search[search]
S_Analytics[analytics]
S_Notify[notification<br><font size=2>worker pool, mock email/push</font>]

%% Определение баз данных и шины данных
DB_User[(PG)]
DB_Catalog[(PG)]
DB_Order[(PG)]
DB_Inventory[(PG)]
DB_Payment[(PG)]
DB_Search[(Elastic)]
DB_Analytics[(ClickHouse)]
Kafka[[Kafka]]

%% Верхнеуровневые связи
Client -->|REST/JSON JWT| Gateway

%% Связи от Gateway к сервисам
Gateway -->|gRPC| S_User
Gateway -->|gRPC| S_Catalog
Gateway -->|gRPC| S_Order
Gateway -->|gRPC| S_Inventory
Gateway -->|gRPC| S_Payment
Gateway -->|gRPC| S_Search
Gateway -->|gRPC| S_Analytics

%% Связи сервисов с их БД
S_User --> DB_User
S_Catalog --> DB_Catalog
S_Order --> DB_Order
S_Inventory --> DB_Inventory
S_Payment --> DB_Payment
S_Search --> DB_Search
S_Analytics --> DB_Analytics

%% События Outbox в Kafka
DB_Catalog & DB_Order & DB_Inventory & DB_Payment -->|outbox → Kafka| Kafka

%% Стриминг из Kafka в поисковые/аналитические движки
Kafka --> DB_Search
Kafka --> DB_Analytics

%% Консьюмер уведомлений
Kafka --> S_Notify
  
```

##  Базовые правила архитектуры

* ️ **Database-per-Service** — у каждого сервиса своя база данных. Чужие таблицы никто не читает, прямой доступ к чужим БД строго запрещен.
*  **Синхронное взаимодействие** — только через `gRPC` и **исключительно для запросов на чтение** (например, `gateway` → `catalog`).
*  **Изменения состояния** — происходят строго **асинхронно** через `Kafka`. Никаких синхронных цепочек на запись.
*  **Гарантия доставки** — события публикуются в брокер только через паттерн **Transactional Outbox** (запись в БД и отправка в очередь в рамках одной транзакции).

# 2. Главный сценарий
```mermaid
sequenceDiagram
    autonumber
    actor Client as Клиент
    participant Order as Order
    participant Kafka as Kafka
    participant Inv as Inventory
    participant Pay as Payment

    Client->>Order: POST /orders (Idempotency-Key)
    Note over Order: TX: Order PENDING<br/>+ outbox: order.created
    Order-->>Client: 202 Accepted (order_id)

    Order->>Kafka: order.created
    Kafka->>Inv: order.created
    Note over Inv: TX: inbox check<br/>резерв всех позиций (all-or-nothing)<br/>Reservation ACTIVE, ttl=15m<br/>+ outbox

    alt Резерв успешен
        Inv->>Kafka: stock.reserved
        Kafka->>Order: stock.reserved
        Note over Order: TX: inbox check<br/>PENDING → RESERVED<br/>+ outbox: payment.requested(amount)
        Order->>Kafka: payment.requested
        Kafka->>Pay: payment.requested
        Note over Pay: Payment PENDING → вызов провайдера<br/>(idempotency key = order_id)<br/>→ TX: результат + outbox

        alt Оплата успешна
            Pay->>Kafka: payment.succeeded
            Kafka->>Order: payment.succeeded
            Note over Order: RESERVED → CONFIRMED<br/>+ outbox: order.confirmed
            Order->>Kafka: order.confirmed
            Kafka->>Inv: order.confirmed
            Note over Inv: Reservation → COMMITTED<br/>(если резерв истёк → stock.commit_failed → refund)
        else Оплата неуспешна
            Pay->>Kafka: payment.failed
            Kafka->>Order: payment.failed
            Note over Order: RESERVED → CANCELLED<br/>+ outbox: order.cancelled(reason=payment)
            Order->>Kafka: order.cancelled
            Kafka->>Inv: order.cancelled
            Note over Inv: Reservation → RELEASED (компенсация)
        end
    else Нет товара
        Inv->>Kafka: stock.reservation_failed
        Kafka->>Order: stock.reservation_failed
        Note over Order: PENDING → CANCELLED<br/>+ outbox: order.cancelled(reason=out_of_stock)
        Order->>Kafka: order.cancelled
    end

    Client->>Order: GET /orders/{id}
    Order-->>Client: status

    Note over Order,Inv: Фоновые воркеры: таймаут заказа в Order,<br/>истечение TTL резерва в Inventory, reconciliation в Payment
```
