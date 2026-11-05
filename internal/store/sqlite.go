package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"kopilka/internal/service"
)

// Store owns the connection to the local SQLite ledger.
type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := migrate(db, migrations); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Load() (service.Ledger, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return service.Ledger{}, err
	}
	defer tx.Rollback()

	l := service.Ledger{Version: service.LedgerVersion}
	if err := tx.QueryRow(`
		SELECT next_category_id, next_operation_id FROM ledger_meta WHERE id = 1
	`).Scan(&l.NextCategoryID, &l.NextOperationID); err != nil {
		return service.Ledger{}, err
	}
	if err := scanRows(tx, "SELECT id, name, archived FROM categories ORDER BY id", func(row *sql.Rows) error {
		var c service.Category
		if err := row.Scan(&c.ID, &c.Name, &c.Archived); err != nil {
			return err
		}
		l.Categories = append(l.Categories, c)
		return nil
	}); err != nil {
		return service.Ledger{}, err
	}
	if err := scanRows(tx, `
		SELECT journal_id, id, date, currency, amount, coalesce(received_currency, ''),
		       coalesce(received_amount, 0), coalesce(category_id, 0), coalesce(rate, 0), coalesce(rubles, 0)
		FROM operations ORDER BY id
	`, func(row *sql.Rows) error {
		var op service.Operation
		if err := row.Scan(&op.JournalID, &op.ID, &op.Date, &op.Currency, &op.Amount, &op.Received,
			&op.Incoming, &op.CategoryID, &op.Rate, &op.Rubles); err != nil {
			return err
		}
		l.Operations = append(l.Operations, op)
		return nil
	}); err != nil {
		return service.Ledger{}, err
	}
	l.Journals, err = loadJournals(tx)
	if err != nil {
		return service.Ledger{}, err
	}
	l.Rates, err = readAmounts(tx, "SELECT currency, rate FROM rates")
	if err != nil {
		return service.Ledger{}, err
	}
	return l, tx.Commit()
}

func scanRows(tx *sql.Tx, query string, scan func(*sql.Rows) error) error {
	rows, err := tx.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func readAmounts(tx *sql.Tx, query string) (map[string]int64, error) {
	var amounts map[string]int64
	err := scanRows(tx, query, func(row *sql.Rows) error {
		var currency string
		var amount int64
		if err := row.Scan(&currency, &amount); err != nil {
			return err
		}
		if amounts == nil {
			amounts = make(map[string]int64)
		}
		amounts[currency] = amount
		return nil
	})
	return amounts, err
}

func (s *Store) Save(l service.Ledger) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := saveSnapshot(tx, l); err != nil {
		return err
	}
	return tx.Commit()
}

// Save the same complete snapshot the UI edits in memory, as one transaction.
func saveSnapshot(tx *sql.Tx, l service.Ledger) error {
	if _, err := tx.Exec(`
		DELETE FROM operations;
		DELETE FROM categories;
		DELETE FROM journal_opening_balances;
		DELETE FROM journals;
		DELETE FROM rates;
	`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO ledger_meta (id, next_category_id, next_operation_id) VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET next_category_id=excluded.next_category_id,
			next_operation_id=excluded.next_operation_id
	`, l.NextCategoryID, l.NextOperationID); err != nil {
		return err
	}
	if err := saveCategories(tx, l.Categories); err != nil {
		return err
	}
	if err := saveJournals(tx, l.Journals); err != nil {
		return err
	}
	for currency, rate := range l.Rates {
		if _, err := tx.Exec("INSERT INTO rates (currency, rate) VALUES (?, ?)", currency, rate); err != nil {
			return err
		}
	}
	stmt, err := tx.Prepare(`
		INSERT INTO operations
			(journal_id, id, date, currency, amount, received_currency, received_amount, category_id, rate, rubles)
		VALUES (?, ?, ?, ?, ?, nullif(?, ''), nullif(?, 0), nullif(?, 0), nullif(?, 0), CASE WHEN ? = '' THEN ? END)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, op := range l.Operations {
		if _, err := stmt.Exec(op.JournalID, op.ID, op.Date, op.Currency, op.Amount, op.Received, op.Incoming,
			op.CategoryID, op.Rate, op.Received, op.Rubles); err != nil {
			return err
		}
	}
	return nil
}

func saveCategories(tx *sql.Tx, categories []service.Category) error {
	for _, c := range categories {
		if _, err := tx.Exec("INSERT INTO categories (id, name, archived) VALUES (?, ?, ?)", c.ID, c.Name, c.Archived); err != nil {
			return err
		}
	}
	return nil
}
