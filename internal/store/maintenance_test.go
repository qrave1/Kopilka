package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExportDBAndReset(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	l, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	l, err = l.WithTrip("Таиланд", "THB")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(l); err != nil {
		t.Fatal(err)
	}
	var backup bytes.Buffer
	if err := s.ExportDB(&backup); err != nil {
		t.Fatal(err)
	}
	reset, err := s.Reset()
	if err != nil || len(reset.Journals) != 0 || len(reset.Categories) != 14 {
		t.Fatalf("reset: %+v, %v", reset, err)
	}
	path := filepath.Join(dir, "backup.db")
	if err := os.WriteFile(path, backup.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	copy, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer copy.Close()
	saved, err := copy.Load()
	if err != nil || len(saved.Journals) != 1 || saved.Journals[0].Name != "Таиланд" {
		t.Fatalf("backup: %+v, %v", saved, err)
	}
}
