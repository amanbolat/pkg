package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	_ "unsafe"

	pkgdecimal "github.com/amanbolat/pkg/decimal"
	pkgpostgres "github.com/amanbolat/pkg/postgres"
	pkgsql "github.com/amanbolat/pkg/sql"
	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/postgres"
	"github.com/go-jet/jet/v2/generator/template"
	jetpg "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	pgtestctr "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("failed to run gengojet generator", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	scheme := flag.String("scheme", "public", "postgres scheme to use for generation")
	dir := flag.String("dir", "./gen/jet", "destination dir for go-jet generated output")
	migrationsPath := flag.String("migrations", "./migrations", "path to database migrations")
	tablesToSkipParam := flag.String("skiptables", "", "tables to skip. Tables that will be skipped during the generation")
	viewsToSkipParam := flag.String("skipviews", "", "views to skip. Views that will be skipped during the generation")
	migrationFormatStr := flag.String("format", "flyway", "migration file format (flyway or gomigrate)")
	flag.Parse()

	migrationFormat, err := pkgsql.ParseMigrationFormat(*migrationFormatStr)
	if err != nil {
		return fmt.Errorf("failed to parse migration format: %w", err)
	}

	tablesToSkipArr := strings.Split(*tablesToSkipParam, ",")
	tablesToSkip := make(map[string]struct{})

	for _, table := range tablesToSkipArr {
		tablesToSkip[table] = struct{}{}
	}

	viewsToSkipArr := strings.Split(*viewsToSkipParam, ",")
	viewsToSkip := make(map[string]struct{})

	for _, view := range viewsToSkipArr {
		viewsToSkip[view] = struct{}{}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	slog.Info("starting postgres container")

	pgCtr, err := pgtestctr.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
	)
	if err != nil {
		return fmt.Errorf("failed to run postgres container: %w", err)
	}

	defer func() {
		err := pgCtr.Terminate(ctx)
		if err != nil {
			slog.Error("failed to terminate postgres container", slog.Any("error", err))
		}
	}()

	dsn, err := pgCtr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	pgConn, err := pkgpostgres.NewSQLConn(ctx, pkgpostgres.SQLConnConfig{
		DSN: dsn,
	})
	if err != nil {
		return fmt.Errorf("failed to create postgres connection: %w", err)
	}

	err = pgConn.Close()
	if err != nil {
		return fmt.Errorf("failed to close postgres connection: %w", err)
	}

	slog.Info("postgres container started", slog.String("dsn", dsn))

	migrator, err := pkgsql.NewMigrator(pkgsql.MigratorConfig{
		MigrationsDir: *migrationsPath,
		DSN:           dsn,
		Format:        migrationFormat,
	})
	if err != nil {
		return fmt.Errorf("failed to create  migrator: %w", err)
	}

	err = migrator.MigrateUp()
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	err = os.RemoveAll(*dir)
	if err != nil {
		return fmt.Errorf("failed to remove previously generated files in the destinations dir %s: %w", *dir, err)
	}

	err = postgres.GenerateDSN(dsn, *scheme, *dir,
		template.Default(jetpg.Dialect).UseSchema(func(schemaMeta metadata.Schema) template.Schema {
			return template.DefaultSchema(schemaMeta).
				UseModel(template.DefaultModel().
					UseView(func(view metadata.Table) template.TableModel {
						if _, ok := viewsToSkip[view.Name]; ok {
							viewModel := template.DefaultViewModel(view)
							viewModel.Skip = true

							return viewModel
						}

						return template.DefaultViewModel(view)
					}).
					UseTable(func(table metadata.Table) template.TableModel {
						if _, ok := tablesToSkip[table.Name]; ok {
							tblModel := template.DefaultTableModel(table)
							tblModel.Skip = true

							return tblModel
						}

						return template.DefaultTableModel(table).
							UseField(func(column metadata.Column) template.TableModelField {
								defaultTableModelField := template.DefaultTableModelField(column).UseTags(
									fmt.Sprintf(`db:"%s.%s"`, table.Name, column.Name),
								)

								if column.DataType.Name == "json" {
									defaultTableModelField.Type = template.NewType(pgtype.JSONCodec{})
								}

								if column.DataType.Kind == metadata.ArrayType && column.DataType.Name == "text" {
									if column.IsNullable {
										defaultTableModelField.Type = template.NewType(&pgtype.Array[string]{})
									} else {
										defaultTableModelField.Type = template.NewType(pgtype.Array[string]{})
									}
								}

								if column.DataType.Name == "numeric" {
									if column.IsNullable {
										defaultTableModelField.Type = template.NewType(&pkgdecimal.Decimal{})
									} else {
										defaultTableModelField.Type = template.NewType(pkgdecimal.Decimal{})
									}
								}

								return defaultTableModelField
							})
					}),
				).
				UseSQLBuilder(template.DefaultSQLBuilder().
					UseView(func(viewMeta metadata.Table) template.ViewSQLBuilder {
						viewBuilder := template.DefaultViewSQLBuilder(viewMeta)
						if _, ok := viewsToSkip[viewMeta.Name]; ok {
							viewBuilder.Skip = true
						}

						return viewBuilder
					}).
					UseTable(func(tableMeta metadata.Table) template.TableSQLBuilder {
						tableBuilder := template.DefaultTableSQLBuilder(tableMeta)
						if _, ok := tablesToSkip[tableMeta.Name]; ok {
							tableBuilder.Skip = true
						}

						return tableBuilder
					}))
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate go-jet files: %w", err)
	}

	slog.Info("successfully generated go-jet files")

	return nil
}
