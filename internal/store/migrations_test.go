package store

import (
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"kopilka/internal/service"
)

func TestMigrateExistingLedgerAndRetryFailedStep(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	l, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	op := service.Operation{Date: "2026-11-05", Currency: "RUB", Amount: 2500,
		CategoryID: l.Categories[0].ID, Rate: service.RateScale, Rubles: 2500}
	if err := s.Save(l.WithOperation(-1, op)); err != nil {
		t.Fatal(err)
	}

	initial, err := fs.ReadFile(migrations, "migrations/001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{
		"migrations/001_init.sql": {Data: initial},
		"migrations/002_note.sql": {Data: []byte("ALTER TABLE operations ADD COLUMN note TEXT")},
		"migrations/003_tag.sql":  {Data: []byte("ALTER TABLE operations ADD COLUMN tag TEXT; INVALID SQL")},
	}
	if err := migrate(s.db, files); err == nil {
		t.Fatal("expected migration 003 to fail")
	}
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 2 {
		t.Fatalf("version after failed migration = %d, error = %v", version, err)
	}

	// Adding the same column on retry succeeds only if the failed step rolled back.
	files["migrations/003_tag.sql"].Data = []byte("ALTER TABLE operations ADD COLUMN tag TEXT")
	if err := migrate(s.db, files); err != nil {
		t.Fatal(err)
	}
	if err := migrate(s.db, files); err != nil {
		t.Fatalf("migrating a second time: %v", err)
	}
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 3 {
		t.Fatalf("version after retry = %d, error = %v", version, err)
	}
	got, err := s.Load()
	if err != nil || len(got.Operations) != 1 || got.Operations[0].Amount != op.Amount {
		t.Fatalf("migration changed journal: %+v, error = %v", got, err)
	}
}
