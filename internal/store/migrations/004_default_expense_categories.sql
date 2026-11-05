WITH presets(position, name, icon) AS (
    VALUES
        (1, 'Счета', 'bills'),
        (2, 'Одежда', 'clothes'),
        (3, 'Образование', 'education'),
        (4, 'Развлечения', 'star'),
        (5, 'Фитнес', 'fitness'),
        (6, 'Продукты', 'food'),
        (7, 'Подарки', 'gift'),
        (8, 'Здоровье', 'health'),
        (9, 'Мебель', 'furniture'),
        (10, 'Домашние животные', 'pet'),
        (11, 'Покупки', 'shopping'),
        (12, 'Поездки', 'transport'),
        (13, 'Путешествие', 'trip'),
        (14, 'Другое', 'tag')
), missing AS (
    SELECT position, name, icon
    FROM presets
    WHERE NOT EXISTS (SELECT 1 FROM categories WHERE categories.name = presets.name)
)
INSERT INTO categories (id, name, archived, icon)
SELECT ledger_meta.next_category_id + row_number() OVER (ORDER BY missing.position) - 1,
       missing.name, 0, missing.icon
FROM missing CROSS JOIN ledger_meta
WHERE ledger_meta.id = 1
ORDER BY missing.position;

UPDATE ledger_meta
SET next_category_id = max(next_category_id, (SELECT coalesce(max(id), 0) + 1 FROM categories))
WHERE id = 1;
