package db

import (
	"database/sql"
	"embed"
	"log"
	"sort"
	"strings"

	"asistente/pkg/domain"
)

//go:embed migrations/sqlite/*.up.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.up.sql
var postgresMigrations embed.FS

func RunMigrations(db *sql.DB, backend string) error {
	if err := ensureMigrationsTable(db, backend); err != nil {
		return domain.Wrapf(domain.ErrMigrateTable, err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return domain.Wrapf(domain.ErrMigrateRead, err)
	}

	var fs embed.FS
	var dir string
	switch backend {
	case "postgres":
		fs = postgresMigrations
		dir = "migrations/postgres"
	default:
		fs = sqliteMigrations
		dir = "migrations/sqlite"
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return domain.Wrapf(domain.ErrMigrateRead, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := fs.ReadFile(dir + "/" + file)
		if err != nil {
			return domain.Wrapf(domain.ErrMigrateRead, err)
		}

		log.Printf("migrator: applying %s", file)

		if _, err := db.Exec(string(content)); err != nil {
			return domain.Wrap(domain.ErrMigrateApply, file+": "+err.Error())
		}

		if err := recordMigration(db, version, backend); err != nil {
			return domain.Wrapf(domain.ErrMigrateRecord, err)
		}
	}

	return nil
}

func ensureMigrationsTable(d *sql.DB, backend string) error {
	var stmt string
	switch backend {
	case "postgres":
		stmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)`
	default:
		stmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	}
	_, err := d.Exec(stmt)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			continue
		}
		applied[v] = true
	}
	return applied, nil
}

func recordMigration(d *sql.DB, version, backend string) error {
	switch backend {
	case "postgres":
		_, err := d.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		return err
	default:
		_, err := d.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		return err
	}
}
