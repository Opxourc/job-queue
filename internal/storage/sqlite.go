package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	ErrUnableToOpenDatabase = errors.New("unable to open database")
	ErrPingToDatabaseFailed = errors.New("ping to database failed")
	ErrInvalidDatabaseFile  = errors.New("database file must be a filename")
	ErrProjectRootNotFound  = errors.New("project root containing go.mod not found")
	ErrQueryFailed          = errors.New("query failed to complete")
	ErrJobAlreadyExists     = errors.New("job already exists")
	ErrJobDoesNotExist      = errors.New("job doesn't exist")
	ErrInvalidJobProvided   = errors.New("provided job is invalid")
)

// ConnectToDatabase opens a database file under the project's data directory.
func ConnectToDatabase(driver, fileName string) (*sqlx.DB, error) {
	root, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	location, err := FindDatabase(root, fileName)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(location), 0o755); err != nil {
		return nil, fmt.Errorf("%w: create data directory: %w", ErrUnableToOpenDatabase, err)
	}

	db, err := sqlx.Open(driver, location)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnableToOpenDatabase, err)
	}

	// Ping to ensure that a connection was truly established
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("%w: %w", ErrPingToDatabaseFailed, err),
				fmt.Errorf("close database after failed ping: %w", closeErr),
			)
		}
		return nil, fmt.Errorf("%w: %w", ErrPingToDatabaseFailed, err)
	}

	log.Printf("Database connection successfully established! Driver: %s", driver)
	return db, nil
}

// FindProjectRoot will attempt to travel upwards in the file tree to find the root of the project.
// The root of such project will be known when `go.mod` exists in the current working directory during the traversal.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working directory: %s", err)
	}

	for {
		info, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil && !info.IsDir() {
			return dir, nil
		}

		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("unrelated error to file existence occured: %s in %s", err, dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrProjectRootNotFound
		}

		dir = parent
	}
}

// FindDatabase will take the provided root directory and attempt to find the database file in it's standardized location.
func FindDatabase(projectRoot, fileName string) (string, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" ||
		filepath.IsAbs(fileName) ||
		filepath.Base(fileName) != fileName ||
		fileName == "." ||
		fileName == ".." {
		return "", ErrInvalidDatabaseFile
	}

	return filepath.Join(projectRoot, "data", fileName), nil
}
