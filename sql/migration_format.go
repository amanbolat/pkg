package pkgsql

// MigrationFormat is a format used to create the migration files.
//
//   - MigrationFileStyleGoMigrate uses the format:
//     {version}_{title}.up.{extension}
//     {version}_{title}.down.{extension}
//
//   - MigrationFileStyleFlyway uses the format:
//     {V}{version}__{migration_name}.sql – for Up migrations.
//     {U}{version}__{migration_name}.sql – for down migrations.
//
/*
ENUM(
flyway
gomigrate
)
*/
type MigrationFormat int8
