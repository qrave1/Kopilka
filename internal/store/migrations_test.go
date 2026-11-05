package store

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestMigrateExistingLedgerAndRetryFailedStep(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "ledger.db")+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &Store{db: db}
	initial, err := fs.ReadFile(migrations, "migrations/001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	journals, err := fs.ReadFile(migrations, "migrations/002_journals.sql")
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{"migrations/001_init.sql": {Data: initial}}
	if err := migrate(s.db, files); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO categories (id, name, archived) VALUES (1, 'Еда', 0);
		UPDATE ledger_meta SET next_category_id = 2;
		INSERT INTO operations (id, date, currency, amount, category_id, rate, rubles)
		VALUES (1, '2026-11-05', 'RUB', 2500, 1, 1000000, 2500);
		UPDATE ledger_meta SET next_operation_id = 2;
		INSERT INTO opening_balances (currency, amount) VALUES ('RUB', 5000);
	`); err != nil {
		t.Fatal(err)
	}
	files["migrations/002_journals.sql"] = &fstest.MapFile{Data: journals}
	icons, err := fs.ReadFile(migrations, "migrations/003_category_icons.sql")
	if err != nil {
		t.Fatal(err)
	}
	files["migrations/003_category_icons.sql"] = &fstest.MapFile{Data: icons}
	presets, err := fs.ReadFile(migrations, "migrations/004_default_expense_categories.sql")
	if err != nil {
		t.Fatal(err)
	}
	files["migrations/004_default_expense_categories.sql"] = &fstest.MapFile{Data: presets}
	files["migrations/005_tag.sql"] = &fstest.MapFile{Data: []byte("ALTER TABLE operations ADD COLUMN tag TEXT; INVALID SQL")}
	if err := migrate(s.db, files); err == nil {
		t.Fatal("expected migration 005 to fail")
	}
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 4 {
		t.Fatalf("version after failed migration = %d, error = %v", version, err)
	}

	// Adding the same column on retry succeeds only if the failed step rolled back.
	files["migrations/005_tag.sql"].Data = []byte("ALTER TABLE operations ADD COLUMN tag TEXT")
	if err := migrate(s.db, files); err != nil {
		t.Fatal(err)
	}
	if err := migrate(s.db, files); err != nil {
		t.Fatalf("migrating a second time: %v", err)
	}
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 5 {
		t.Fatalf("version after retry = %d, error = %v", version, err)
	}
	got, err := s.Load()
	if err != nil || len(got.Operations) != 1 || got.Operations[0].Amount != 2500 || got.Operations[0].JournalID != 1 || got.Balances(1)["RUB"] != 2500 {
		t.Fatalf("migration changed journal: %+v, error = %v", got, err)
	}
	if icon := got.Category(1).Icon; icon != "food" {
		t.Fatalf("default food category icon = %q", icon)
	}
}
