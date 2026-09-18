package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPostgreSQLIntegrationAppliesCoreMigrations(t *testing.T) {
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

	if err := resetPublicSchema(ctx, database); err != nil {
		t.Fatalf("reset test schema: %v", err)
	}
	if err := database.Probe(ctx); !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("expected readiness probe to fail before migration, got %v", err)
	}

	if err := applyMigration(ctx, database, "000001_core.up.sql"); err != nil {
		t.Fatalf("apply core up migration: %v", err)
	}
	if err := database.Probe(ctx); !errors.Is(err, ErrProbeFailed) {
		t.Fatalf("expected readiness probe to fail before upload-parts migration, got %v", err)
	}

	if err := applyMigration(ctx, database, "000002_upload_parts.up.sql"); err != nil {
		t.Fatalf("apply upload-parts up migration: %v", err)
	}
	if err := database.Probe(ctx); err != nil {
		t.Fatalf("probe PostgreSQL after migrations: %v", err)
	}

	for _, table := range []string{
		"libraries",
		"library_memberships",
		"original_objects",
		"assets",
		"device_asset_states",
		"upload_sessions",
		"upload_parts",
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

	if err := applyMigration(ctx, database, "000002_upload_parts.down.sql"); err != nil {
		t.Fatalf("apply upload-parts down migration: %v", err)
	}
	if err := applyMigration(ctx, database, "000001_core.down.sql"); err != nil {
		t.Fatalf("apply core down migration: %v", err)
	}

	var remaining int
	if err := database.pool.QueryRow(ctx, "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('assets', 'upload_parts')").Scan(&remaining); err != nil {
		t.Fatalf("verify down migrations: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected migrated tables to be removed, found %d", remaining)
	}
}

func resetAndApplyMigrations(ctx context.Context, database *PostgreSQL) error {
	if err := resetPublicSchema(ctx, database); err != nil {
		return err
	}
	for _, name := range []string{"000001_core.up.sql", "000002_upload_parts.up.sql"} {
		if err := applyMigration(ctx, database, name); err != nil {
			return err
		}
	}
	return nil
}

func resetPublicSchema(ctx context.Context, database *PostgreSQL) error {
	return execSQLScript(ctx, database, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
}

func applyMigration(ctx context.Context, database *PostgreSQL, name string) error {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	if err != nil {
		return err
	}
	return execSQLScript(ctx, database, string(content))
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
