package pkgsql

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // import pgx driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // import pgx driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/afero"
)

const goMigrateTimeFormat = "20060102150405"

// Regex to match the flyway migration file format.
//
// Example:
// * U1683211973__cdr_tables.sql
// * V1683211973__cdr_tables.sql
var flywayFileFormatRegex = regexp.MustCompile(`^([UV])(\d+)__(\w+)\.sql$`)

type MigratorConfig struct {
	// MigrationsDir is a path to the migrations directory. If MigrationsFs is nil, the Migrator will use it directly.
	// If MigrationsFs is not nil, the path in MigrationsDir is used to strip the prefix.
	//
	// Example:
	//
	//     //go:embed migrations/*
	//     var MigrationsFS embed.FS
	//
	// Here, 'migrations' is a prefix that should be stripped from the path.
	// This allows the iofs.Driver to read its content correctly.
	MigrationsDir string
	MigrationsFs  fs.FS
	DSN           string
	Format        MigrationFormat
}

// Migrator is Postgres database schem migrator.
type Migrator struct {
	migrator *migrate.Migrate
}

// NewMigrator returns a new Migrator.
func NewMigrator(cfg MigratorConfig) (*Migrator, error) {
	var goMigrator *migrate.Migrate
	switch cfg.Format {
	case MigrationFormatGomigrate:
		if cfg.MigrationsFs != nil {
			fsDriver, err := iofs.New(cfg.MigrationsFs, cfg.MigrationsDir)
			if err != nil {
				return nil, fmt.Errorf("failed to create fs driver: %w", err)
			}

			goMigrator, err = migrate.NewWithSourceInstance("iofs", fsDriver, cfg.DSN)
			if err != nil {
				return nil, err
			}
		} else {
			migrations, err := filepath.Abs(cfg.MigrationsDir)
			if err != nil {
				return nil, err
			}

			goMigrator, err = migrate.New(fmt.Sprintf("file://%v", migrations), cfg.DSN)
			if err != nil {
				return nil, err
			}
		}
	case MigrationFormatFlyway:
		memFs := afero.NewMemMapFs()

		err := filepath.WalkDir(cfg.MigrationsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			tmpFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer tmpFile.Close()

			goMigrateName, err := flywayFormatToGoMigrate(d.Name())
			if err != nil {
				return err
			}

			goMigrateFile, err := memFs.Create(goMigrateName)
			if err != nil {
				return err
			}
			defer goMigrateFile.Close()

			_, err = io.Copy(goMigrateFile, tmpFile)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		migrateIOFS, err := iofs.New(afero.NewIOFS(memFs), "")
		if err != nil {
			return nil, err
		}

		goMigrator, err = migrate.NewWithSourceInstance("iofs", migrateIOFS, cfg.DSN)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown migration file format")
	}

	pm := &Migrator{migrator: goMigrator}

	return pm, nil
}

// MigrateUp applies up migrations.
func (m *Migrator) MigrateUp() error {
	return m.migrator.Up()
}

// MigrateDown applies down migrations.
func (m *Migrator) MigrateDown() error {
	return m.migrator.Down()
}

// flywayFormatToGoMigrate converts SQL migration file name from
// Flyway format to go-migrate format.
//
// Example:
//
//	V1683211973_tables.sql -> 20230504225253_tables.up.sql
//
// V = up
// U = down
func flywayFormatToGoMigrate(name string) (string, error) {
	match := flywayFileFormatRegex.FindStringSubmatch(name)
	if len(match) < 4 {
		return "", fmt.Errorf("file %s doesnt match the regex: %v", name, flywayFileFormatRegex.String())
	}

	var isUpMigration bool
	if match[1] == "V" {
		isUpMigration = true
	}

	// The version must be manually converted, otherwise, during the down migrations,
	// gomigrate will not be able to correctly identify the file name.
	intVersion, err := strconv.Atoi(match[2])
	if err != nil {
		return "", fmt.Errorf("failed to convert %s to unix seconds: %w", match[2], err)
	}
	formattedVersion := time.Unix(int64(intVersion), 0).Format(goMigrateTimeFormat)

	sb := strings.Builder{}
	sb.WriteString(formattedVersion)
	sb.WriteString("_")
	sb.WriteString(match[3])
	sb.WriteString(".")
	if isUpMigration {
		sb.WriteString("up")
	} else {
		sb.WriteString("down")
	}
	sb.WriteString(".")
	sb.WriteString("sql")

	return sb.String(), nil
}
