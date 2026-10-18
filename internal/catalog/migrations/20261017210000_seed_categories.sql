-- +goose Up
-- Корневые категории
INSERT INTO category (name, slug)
VALUES ('Одежда', 'clothing'),
       ('Дом и сад', 'home'),
       ('Книги', 'books'),
       ('Спорт и отдых', 'sports');

-- Второй уровень
INSERT INTO category (name, slug, parent_id)
SELECT v.name, v.slug, p.id
FROM (VALUES ('Аудиотехника', 'audio', 'electronics'),
             ('Телевизоры', 'tv', 'electronics'),
             ('Ноутбуки', 'laptops', 'computers'),
             ('Мониторы', 'monitors', 'computers'),
             ('Мужская одежда', 'clothing-men', 'clothing'),
             ('Женская одежда', 'clothing-women', 'clothing'),
             ('Мебель', 'furniture', 'home'),
             ('Кухня', 'kitchen', 'home'),
             ('Художественная литература', 'fiction', 'books'),
             ('Программирование', 'programming-books', 'books'),
             ('Велосипеды', 'bicycles', 'sports'),
             ('Туризм', 'camping', 'sports')) AS v(name, slug, parent_slug)
JOIN category p ON p.slug = v.parent_slug;

-- +goose Down
-- Сначала дети, потом родители (иначе FK category_parent_id_fkey)
DELETE FROM category
WHERE slug IN ('audio', 'tv', 'laptops', 'monitors',
               'clothing-men', 'clothing-women', 'furniture', 'kitchen',
               'fiction', 'programming-books', 'bicycles', 'camping');

DELETE FROM category
WHERE slug IN ('clothing', 'home', 'books', 'sports');
