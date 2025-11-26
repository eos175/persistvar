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
