package storage_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/eos175/persistvar"
	"github.com/eos175/persistvar/storage"
)

// runStorageTests is a generic suite to test any Storage implementation.
func runStorageTests(t *testing.T, s persistvar.Storage) {
	key := "testkey"
	val := []byte("hello")
	val2 := []byte("world")

	// 1. Test Save and Load
	if err := s.Save(key, val, nil); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := s.Load(key)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !bytes.Equal(loaded, val) {
		t.Errorf("Expected %s, got %s", val, loaded)
	}

	// 2. Test Update (with old value)
	if err := s.Save(key, val2, val); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	loaded2, err := s.Load(key)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !bytes.Equal(loaded2, val2) {
		t.Errorf("Expected %s, got %s", val2, loaded2)
	}

	// 3. Test Delete
	if err := s.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = s.Load(key)
	if err != persistvar.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	// 4. Test Close
	if err := s.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestFileStorage(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "test_file_storage")
	defer os.RemoveAll(tmpDir)

	s, err := storage.NewFileStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	runStorageTests(t, s)
}
