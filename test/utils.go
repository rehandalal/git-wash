package test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Get the path for the golden test data file
func TestDataPath(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get caller information")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata")

	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			t.Fatal("could not make test data directory")
		}
	}

	return filepath.Join(path, t.Name()+".golden")
}

// Write a golden test data file
func WriteTestData(t *testing.T, content []byte) {
	t.Helper()

	path := TestDataPath(t)

	_, err := os.Stat(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal("could not make test data directory")
		}
	}

	err = os.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// Read a golden test data file
func ReadTestData(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile(TestDataPath(t))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}
