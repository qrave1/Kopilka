package store

import (
	"database/sql"
	"fmt"

	"kopilka/internal/service"
)

func loadJournals(tx *sql.Tx) ([]service.Journal, error) {
	var journals []service.Journal
	indices := make(map[int64]int)
	err := scanRows(tx, "SELECT id, kind, name, currency, archived, started FROM journals ORDER BY id", func(row *sql.Rows) error {
		var j service.Journal
		if err := row.Scan(&j.ID, &j.Kind, &j.Name, &j.Currency, &j.Archived, &j.Started); err != nil {
			return err
		}
		indices[j.ID] = len(journals)
		journals = append(journals, j)
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = scanRows(tx, "SELECT journal_id, currency, amount FROM journal_opening_balances", func(row *sql.Rows) error {
		var id, amount int64
		var currency string
		if err := row.Scan(&id, &currency, &amount); err != nil {
			return err
		}
		index, ok := indices[id]
		if !ok {
			return fmt.Errorf("начальные наличные без журнала %d", id)
		}
		journal := &journals[index]
		if journal.Opening == nil {
			journal.Opening = make(map[string]int64)
		}
		journal.Opening[currency] = amount
		return nil
	})
	return journals, err
}

func saveJournals(tx *sql.Tx, journals []service.Journal) error {
	for _, j := range journals {
		if _, err := tx.Exec(`
			INSERT INTO journals (id, kind, name, currency, archived, started) VALUES (?, ?, ?, ?, ?, ?)
		`, j.ID, j.Kind, j.Name, j.Currency, j.Archived, j.Started); err != nil {
			return err
		}
		for currency, amount := range j.Opening {
			if _, err := tx.Exec(`
				INSERT INTO journal_opening_balances (journal_id, currency, amount) VALUES (?, ?, ?)
			`, j.ID, currency, amount); err != nil {
				return err
			}
		}
	}
	return nil
}
