package store

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"kopilka/internal/service"
)

//go:embed migrations/*.sql
var migrations embed.FS

func migrate(db *sql.DB, files fs.FS) error {
	names, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return fmt.Errorf("миграции базы отсутствуют")
	}
	for i, name := range names {
		base := path.Base(name)
		prefix := fmt.Sprintf("%03d_", i+1)
		if !strings.HasPrefix(base, prefix) || len(base) <= len(prefix)+len(".sql") {
			return fmt.Errorf("неверное имя миграции %s, ожидается префикс %s", base, prefix)
		}
	}

	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version < 0 || version > len(names) {
		return fmt.Errorf("неподдерживаемая версия базы %d", version)
	}
	for i := version; i < len(names); i++ {
		if err := applyMigration(db, files, names[i], i+1); err != nil {
			return fmt.Errorf("миграция %s: %w", names[i], err)
		}
	}
	return nil
}

func applyMigration(db *sql.DB, files fs.FS, name string, version int) error {
	script, err := fs.ReadFile(files, name)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var current int
	if err := tx.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return err
	}
	if current >= version {
		return nil
	}
	if current != version-1 {
		return fmt.Errorf("ожидалась версия %d, получена %d", version-1, current)
	}
	if _, err := tx.Exec(string(script)); err != nil {
		return err
	}
	if version == 1 {
		if err := saveSnapshot(tx, service.NewLedger()); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return err
	}
	return tx.Commit()
}
