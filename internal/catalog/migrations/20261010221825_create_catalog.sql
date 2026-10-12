-- +goose Up
INSERT INTO category (name, slug)
VALUES ('Электроника', 'electronics');

INSERT INTO category (name, slug, parent_id)
VALUES ('Телефоны', 'phones', (SELECT id FROM category WHERE slug = 'electronics'));

INSERT INTO category (name, slug, parent_id)
VALUES ('Компьютеры', 'computers', (SELECT id FROM category WHERE slug = 'electronics'));

INSERT INTO product (id, seller_id, title, description, category_id, price_amount)
VALUES (
        'b3c1e7a2-5d4f-4e8a-9c6b-2f1a0d9e8c71',
        '1ab2f99c-4063-4a53-a226-b03e8020d021',
        'iPhone 18 Pro Max 512 Гб',
        'Новый айфон оснащен новой камерой и процессором. Успейте купить!',
        (SELECT id FROM category WHERE slug = 'phones'),
        100000
       );

INSERT INTO product (id, seller_id, title, description, category_id, price_amount)
VALUES (
           '661899f7-4f7d-4491-9280-52a728d7b4c8',
           'e2f09df1-19ac-459d-9946-143ed7a92df1',
           'Macbook M4 Pro 16 дюймов 512 Гб',
           'Новый Macbook с системой охлаждения и мощным процессором',
           (SELECT id FROM category WHERE slug = 'computers'),
           160000
       );

-- +goose Down
DELETE FROM product WHERE id = '661899f7-4f7d-4491-9280-52a728d7b4c8';
DELETE FROM product WHERE id = 'b3c1e7a2-5d4f-4e8a-9c6b-2f1a0d9e8c71';
DELETE FROM category WHERE slug = 'computers';
DELETE FROM category WHERE slug = 'phones';
DELETE FROM category WHERE slug = 'electronics';
