package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // or your database
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Add this import
)

var MigrationsPath = "migrations"

func RunMigrations(dsn string, up bool) error {
	if _, err := os.Stat(MigrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", MigrationsPath)
	}

	absPath, err := filepath.Abs(MigrationsPath)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", absPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if up {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		return nil
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func Ping(ctx context.Context, dsn string) error {
	_ = ctx
	return nil
}
