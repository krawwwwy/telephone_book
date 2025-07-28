package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"                     // либа для миграций
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер для миграций postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // драйвер для получения миграций из файлов
)

func main() {
	var migrationsPath, migrationsTable string
	var down bool

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations")
	flag.BoolVar(&down, "down", false, "run down migrations")
	flag.Parse()

	dsn := os.Getenv("DSN")
	if dsn == "" {
		panic("DSN is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("%s&x-migrations-table=%s", dsn, migrationsTable),
	)

	v, d, _ := m.Version()
	fmt.Printf("Current version: %d, dirty: %v\n", v, d)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Current version: %d, dirty: %v\n", v, d)

	var migrationErr error
	if down {
		migrationErr = m.Down()

	} else {
		migrationErr = m.Up()
	}

	if migrationErr != nil {
		if errors.Is(migrationErr, migrate.ErrNoChange) {
			if down {
				fmt.Println("no migrations to rollback")
			} else {
				fmt.Println("no migrations to apply")
			}
			return
		}
		panic(migrationErr)
	}

	if down {
		fmt.Println("migrations rolled back successfully")
	} else {
		fmt.Println("migrations applied successfully")
	}
}
