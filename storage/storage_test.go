package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

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

func TestStressConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := storage.NewFileStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	mgr := persistvar.NewVarManager(fs, nil)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr.AutoSync(ctx, 10*time.Millisecond)

	var wg sync.WaitGroup
	numVars := 10
	numGoroutines := 50

	for i := 0; i < numVars; i++ {
		key := fmt.Sprintf("var_%d", i)
		v, _ := persistvar.NewVar(mgr, key, 0)

		for j := 0; j < numGoroutines; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v.UpdateLazy(func(curr int) (int, bool) {
					return curr + 1, true
				})
			}()
		}
	}

	wg.Wait()
	
	// Close mgr to ensure final sync
	if err := mgr.Close(); err != nil {
		t.Errorf("Manager.Close() failed: %v", err)
	}

	// Verify final values by loading from the same directory
	fs2, err := storage.NewFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage for verification: %v", err)
	}
	mgr2 := persistvar.NewVarManager(fs2, nil)
	defer mgr2.Close()

	for i := 0; i < numVars; i++ {
		key := fmt.Sprintf("var_%d", i)
		v, err := persistvar.NewVar(mgr2, key, 0)
		if err != nil {
			t.Fatalf("Failed to reload var %s: %v", key, err)
		}
		if val := v.Get(); val != numGoroutines {
			t.Errorf("Data integrity error for %s: expected %d, got %d", key, numGoroutines, val)
		}
	}
}
