package persistvar_test

import (
	"os"
	"testing"

	"github.com/eos175/persistvar"
	"github.com/eos175/persistvar/storage"
)

func TestNewVar_Singleton(t *testing.T) {
	// 1. Setup Storage
	tmpDir, err := os.MkdirTemp("", "persistvar_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	st, err := storage.NewFileStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	mgr := persistvar.NewVarManager(st)
	defer mgr.Close()

	// 2. Crear primera instancia
	v1, err := persistvar.NewVar(mgr, "counter", 100)
	if err != nil {
		t.Fatalf("Error creating v1: %v", err)
	}

	// 3. Crear segunda instancia con la misma clave
	v2, err := persistvar.NewVar(mgr, "counter", 200) // El valor por defecto 200 debería ser ignorado
	if err != nil {
		t.Fatalf("Error creating v2: %v", err)
	}

	// 4. Verificación 1: Igualdad de punteros
	if v1 != v2 {
		t.Errorf("FAILED: v1 and v2 should be the same pointer instance. Got %p and %p", v1, v2)
	}

	// 5. Verificación 2: Estado compartido
	v1.SetLazy(500)
	if v2.Get() != 500 {
		t.Errorf("FAILED: v2 did not reflect change in v1. Expected 500, got %v", v2.Get())
	}

	// 6. Verificación 3: Tipo incorrecto
	_, err = persistvar.NewVar(mgr, "counter", "string_value")
	if err == nil {
		t.Error("FAILED: Expected error when creating existing var with different type, got nil")
	}
}

func TestVar_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "persistvar_test_update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	st, err := storage.NewFileStorage(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	mgr := persistvar.NewVarManager(st)
	defer mgr.Close()

	v, err := persistvar.NewVar(mgr, "sequence", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Test UpdateLazy (Incremento)
	newVal := v.UpdateLazy(func(curr int) (int, bool) {
		return curr + 1, true
	})
	if newVal != 1 {
		t.Errorf("UpdateLazy failed: expected 1, got %d", newVal)
	}
	if v.Get() != 1 {
		t.Errorf("Get after UpdateLazy failed: expected 1, got %d", v.Get())
	}

	// 2. Test Update (No change optimization)
	// Intentamos actualizar al mismo valor, retornando false
	valAfterNoChange, err := v.Update(func(curr int) (int, bool) {
		return curr, false
	})
	if err != nil {
		t.Fatal(err)
	}
	if valAfterNoChange != 1 {
		t.Errorf("Update with no change failed: expected 1, got %d", valAfterNoChange)
	}

	// 3. Test Update (Persistencia inmediata)
	valPersisted, err := v.Update(func(curr int) (int, bool) {
		return curr + 10, true
	})
	if err != nil {
		t.Fatal(err)
	}
	if valPersisted != 11 {
		t.Errorf("Update persisted failed: expected 11, got %d", valPersisted)
	}

	// Verificar en una nueva instancia que se persistió
	v2, err := persistvar.NewVar(mgr, "sequence", 0)
	if err != nil {
		t.Fatal(err)
	}
	if v2.Get() != 11 {
		t.Errorf("Persistence check failed: expected 11, got %d", v2.Get())
	}
}
