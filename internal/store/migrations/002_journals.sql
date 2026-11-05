CREATE TABLE journals (
    id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('trip', 'everyday')),
    name TEXT NOT NULL,
    currency TEXT NOT NULL CHECK (currency IN ('RUB', 'USD', 'THB')),
    archived INTEGER NOT NULL CHECK (archived IN (0, 1)),
    started INTEGER NOT NULL CHECK (started IN (0, 1))
) STRICT;

INSERT INTO journals (id, kind, name, currency, archived, started)
SELECT 1, 'trip', 'Моя поездка', 'THB', 0, started
FROM ledger_meta
WHERE started = 1 OR EXISTS (SELECT 1 FROM operations) OR EXISTS (SELECT 1 FROM opening_balances);

CREATE TABLE journal_opening_balances (
    journal_id INTEGER NOT NULL REFERENCES journals(id),
    currency TEXT NOT NULL CHECK (currency IN ('RUB', 'USD', 'THB')),
    amount INTEGER NOT NULL CHECK (amount >= 0),
    PRIMARY KEY (journal_id, currency)
) STRICT;

INSERT INTO journal_opening_balances (journal_id, currency, amount)
SELECT 1, currency, amount FROM opening_balances;
DROP TABLE opening_balances;
ALTER TABLE ledger_meta DROP COLUMN started;

CREATE TABLE operations_new (
    journal_id INTEGER NOT NULL REFERENCES journals(id),
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

INSERT INTO operations_new (journal_id, id, date, currency, amount, received_currency, received_amount, category_id, rate, rubles)
SELECT 1, id, date, currency, amount, received_currency, received_amount, category_id, rate, rubles FROM operations;
DROP TABLE operations;
ALTER TABLE operations_new RENAME TO operations;
CREATE INDEX operations_journal_date ON operations(journal_id, date);
CREATE INDEX operations_category_date ON operations(category_id, date);
