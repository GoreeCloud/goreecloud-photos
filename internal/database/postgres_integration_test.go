package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPostgreSQLIntegrationAppliesCoreMigration(t *testing.T) {
	connectionString := os.Getenv("GC_PHOTOS_TEST_DATABASE_URL")
	if connectionString == "" {
		t.Skip("GC_PHOTOS_TEST_DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := NewPostgreSQL(ctx, connectionString)
	if err != nil {
		t.Fatalf("initialize PostgreSQL adapter: %v", err)
	}
	defer database.Close()

	if err := database.Probe(ctx); err != nil {
		t.Fatalf("probe PostgreSQL: %v", err)
	}

	if err := execSQLScript(ctx, database, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("reset test schema: %v", err)
	}

	up, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_core.up.sql"))
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	if err := execSQLScript(ctx, database, string(up)); err != nil {
		t.Fatalf("apply up migration: %v", err)
	}

	for _, table := range []string{
		"libraries",
		"library_memberships",
		"original_objects",
		"assets",
		"device_asset_states",
		"upload_sessions",
		"sync_changes",
		"jobs",
		"idempotency_records",
	} {
		var relation string
		if err := database.pool.QueryRow(ctx, "SELECT to_regclass($1)::text", "public."+table).Scan(&relation); err != nil {
			t.Fatalf("inspect table %s: %v", table, err)
		}
		if relation != table {
			t.Fatalf("expected table %s, got %q", table, relation)
		}
	}

	down, err := os.ReadFile(filepath.Join("..", "..", "migrations", "000001_core.down.sql"))
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if err := execSQLScript(ctx, database, string(down)); err != nil {
		t.Fatalf("apply down migration: %v", err)
	}

	var remaining int
	if err := database.pool.QueryRow(ctx, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'assets'").Scan(&remaining); err != nil {
		t.Fatalf("verify down migration: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected assets table to be removed, found %d", remaining)
	}
}

func execSQLScript(ctx context.Context, database *PostgreSQL, script string) error {
	connection, err := database.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer connection.Release()

	_, err = connection.Conn().PgConn().Exec(ctx, script).ReadAll()
	return err
}
