package database

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewPostgreSQLRejectsEmptyConfiguration(t *testing.T) {
	database, err := NewPostgreSQL(context.Background(), "")
	if database != nil {
		database.Close()
		t.Fatal("expected no database adapter")
	}
	if !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("expected ErrInvalidConfiguration, got %v", err)
	}
}

func TestNewPostgreSQLDoesNotLeakConnectionSecretOnParseFailure(t *testing.T) {
	const secret = "do-not-log-this-password"
	database, err := NewPostgreSQL(context.Background(), "postgres://photos:"+secret+"@%zz")
	if database != nil {
		database.Close()
		t.Fatal("expected no database adapter")
	}
	if !errors.Is(err, ErrInvalidConfiguration) {
		t.Fatalf("expected ErrInvalidConfiguration, got %v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("database configuration error leaked the connection secret")
	}
}
