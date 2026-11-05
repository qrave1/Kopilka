CREATE TABLE ledger_meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    started INTEGER NOT NULL CHECK (started IN (0, 1)),
    next_category_id INTEGER NOT NULL CHECK (next_category_id > 0),
    next_operation_id INTEGER NOT NULL CHECK (next_operation_id > 0)
) STRICT;

CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    archived INTEGER NOT NULL CHECK (archived IN (0, 1))
) STRICT;

CREATE TABLE opening_balances (
    currency TEXT PRIMARY KEY CHECK (currency IN ('RUB', 'USD', 'THB')),
    amount INTEGER NOT NULL CHECK (amount >= 0)
) STRICT;

CREATE TABLE rates (
    currency TEXT PRIMARY KEY CHECK (currency IN ('THB', 'USD')),
    rate INTEGER NOT NULL CHECK (rate BETWEEN 0 AND 1000000000)
) STRICT;

CREATE TABLE operations (
    id INTEGER PRIMARY KEY,
    date TEXT NOT NULL,
    currency TEXT NOT NULL CHECK (currency IN ('RUB', 'USD', 'THB')),
    amount INTEGER NOT NULL CHECK (amount > 0),
    received_currency TEXT CHECK (received_currency IN ('RUB', 'USD', 'THB') AND received_currency <> currency),
    received_amount INTEGER CHECK (received_amount > 0),
    category_id INTEGER REFERENCES categories(id),
    rate INTEGER CHECK (rate > 0),
    rubles INTEGER CHECK (rubles >= 0),
    CHECK (
        (received_currency IS NULL AND received_amount IS NULL
            AND category_id IS NOT NULL AND rate IS NOT NULL AND rubles IS NOT NULL)
        OR
        (received_currency IS NOT NULL AND received_amount IS NOT NULL
            AND category_id IS NULL AND rate IS NULL AND rubles IS NULL)
    )
) STRICT;

CREATE INDEX operations_date ON operations(date);
CREATE INDEX operations_category_date ON operations(category_id, date);
