package pkgpostgres

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresMigrator is Postgres database schem migrator.
type PostgresMigrator struct {
	migrator *migrate.Migrate
}

// NewPostgresMigrator returns a new PostgresMigrator.
// * migrationsPath is the path to the directory containing the migrations.
// * dsn is the Postgres connection string.
func NewPostgresMigrator(migrationsPath string, dsn string) (*PostgresMigrator, error) {
	migrations, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, err
	}

	m, err := migrate.New(fmt.Sprintf("file://%v", migrations), dsn)
	if err != nil {
		return nil, err
	}

	pm := &PostgresMigrator{migrator: m}

	return pm, nil
}

// MigrateUp applies up migrations.
func (m *PostgresMigrator) MigrateUp() error {
	return m.migrator.Up()
}

// MigrateDown applies down migrations.
func (m *PostgresMigrator) MigrateDown() error {
	return m.migrator.Down()
}
