package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Opxourc/job-queue/internal/storage"
)

// TestConnectToDatabaseUsesProjectDataDirectory tests if the database connection can be established.
func TestConnectToDatabaseUsesProjectDataDirectory(t *testing.T) {
	root, err := storage.FindProjectRoot()
	if err != nil {
		t.Fatal(err)
	}

	tempFile, err := os.CreateTemp("", "sqlite-storage-test-*.db")
	if err != nil {
		t.Fatal(err)
	}

	filename := filepath.Base(tempFile.Name())
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(tempPath); err != nil {
		t.Fatal(err)
	}

	databasePath := filepath.Join(root, "data", filename)
	db, err := storage.ConnectToDatabase("sqlite", filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
		if err := os.Remove(databasePath); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove test database: %v", err)
		}
	}()

	var sequence int
	var name, actualPath string
	if err := db.QueryRowx("PRAGMA database_list").Scan(&sequence, &name, &actualPath); err != nil {
		t.Fatal(err)
	}
	if actualPath != databasePath {
		t.Fatalf("database path = %q, want %q", actualPath, databasePath)
	}

	if _, err := db.Exec("CREATE TABLE connection_test (value TEXT)"); err != nil {
		t.Fatalf("create test table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO connection_test (value) VALUES (?)", "connected"); err != nil {
		t.Fatalf("insert test value: %v", err)
	}

	var value string
	if err := db.Get(&value, "SELECT value FROM connection_test"); err != nil {
		t.Fatalf("read test value: %v", err)
	}
	if value != "connected" {
		t.Fatalf("stored value = %q, want %q", value, "connected")
	}
}

// TestValidateDatabaseLocationRejectsPaths tests if database paths that are invalid will fail as intended.
func TestValidateDatabaseLocationRejectsPaths(t *testing.T) {
	for _, filename := range []string{"", ".", "..", "../outside.db", filepath.Join("data", "outside.db")} {
		t.Run(filename, func(t *testing.T) {
			if _, err := storage.FindDatabase(t.TempDir(), filename); err != storage.ErrInvalidDatabaseFile {
				t.Fatalf("validateDatabaseLocation(%q) error = %v, want %v", filename, err, storage.ErrInvalidDatabaseFile)
			}
		})
	}
}
