package store

import (
	"io"
	"os"
	"path/filepath"

	"kopilka/internal/service"
)

func (s *Store) Reset() (service.Ledger, error) {
	presets, err := migrations.ReadFile("migrations/004_default_expense_categories.sql")
	if err != nil {
		return service.Ledger{}, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return service.Ledger{}, err
	}
	defer tx.Rollback()
	if err := saveSnapshot(tx, service.Ledger{NextCategoryID: 1, NextOperationID: 1}); err != nil {
		return service.Ledger{}, err
	}
	if _, err := tx.Exec(string(presets)); err != nil {
		return service.Ledger{}, err
	}
	if err := tx.Commit(); err != nil {
		return service.Ledger{}, err
	}
	return s.Load()
}

func (s *Store) ExportDB(output io.Writer) error {
	dir, err := os.MkdirTemp(s.dir, ".export-*")
	if err != nil {
		return err
	}
	defer os.Remove(dir)
	path := filepath.Join(dir, "ledger.db")
	defer os.Remove(path)
	if _, err := s.db.Exec("VACUUM INTO ?", path); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(output, file)
	return err
}
