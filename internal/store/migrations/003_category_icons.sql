ALTER TABLE categories ADD COLUMN icon TEXT NOT NULL DEFAULT 'tag';

UPDATE categories SET icon = CASE name
    WHEN 'Еда' THEN 'food'
    WHEN 'Транспорт' THEN 'transport'
    WHEN 'Жильё' THEN 'home'
    WHEN 'Развлечения' THEN 'star'
    WHEN 'Покупки' THEN 'shopping'
    ELSE 'tag'
END;
