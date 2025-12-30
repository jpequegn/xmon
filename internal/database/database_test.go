// internal/database/database_test.go
package database

import (
	"os"
	"testing"
)

func TestNewDB(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "xmon-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	db, err := New(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("expected non-nil db")
	}
}
