// Package mysql implementa los puertos de persistencia del dominio sobre MySQL.
package mysql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/go-sql-driver/mysql"

	"kanban/internal/domain/kanban"
	"kanban/internal/domain/shared"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// dbtx lo satisfacen tanto *sql.DB como *sql.Tx.
type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Connect abre el pool y espera a que MySQL responda.
func Connect(dsn string) (*sql.DB, error) {
	d, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(10)
	d.SetConnMaxLifetime(5 * time.Minute)
	for i := 0; i < 30; i++ {
		if err = d.Ping(); err == nil {
			return d, nil
		}
		log.Printf("esperando a MySQL: %v", err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

// Migrate aplica en orden los archivos SQL pendientes y los registra en schema_migrations.
func Migrate(dsn string) error {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return err
	}
	cfg.MultiStatements = true
	conn, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(100) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		var n int
		if err := conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version=?`, name).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(string(body)); err != nil {
			return fmt.Errorf("migración %s: %w", name, err)
		}
		if _, err := conn.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			return err
		}
		log.Printf("migración aplicada: %s", name)
	}
	return nil
}

// mapErr traduce errores de base de datos a errores del dominio.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return shared.ErrNotFound
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1452, 1406: // violación de FK (referencia inexistente) / dato demasiado largo
			return shared.ErrInvalid
		}
	}
	return err
}

func isRetryable(err error) bool {
	var me *mysql.MySQLError
	return errors.Is(err, kanban.ErrConcurrentChange) ||
		(errors.As(err, &me) && (me.Number == 1213 || me.Number == 1205)) // deadlock / lock wait timeout
}

func retry(fn func() error) error {
	var err error
	for i := 0; i < 6; i++ {
		if err = fn(); !isRetryable(err) {
			return err
		}
		time.Sleep(time.Duration(10*(i+1)) * time.Millisecond)
	}
	return err
}

func nilIfZero(i int) any {
	if i == 0 {
		return nil
	}
	return i
}

func ptrArg(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func affected(res sql.Result, err error) error {
	if err != nil {
		return mapErr(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return shared.ErrNotFound
	}
	return nil
}
