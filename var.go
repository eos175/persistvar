package persistvar

import (
	"bytes"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type Var[T any] struct {
	key        string
	value      T
	storage    Storage
	mu         sync.RWMutex
	dirty      bool
	lastSynced []byte // Snapshot del último estado guardado
}

func NewVar[T any](m *VarManager, key string, defaultValue T) (*Var[T], error) {
	v := &Var[T]{key: key, storage: m.storage}

	data, err := m.storage.Load(key)
	if err == nil {
		if unmarshalErr := msgpack.Unmarshal(data, &v.value); unmarshalErr != nil {
			return nil, unmarshalErr
		}
		v.lastSynced = data // Guardamos el snapshot
	} else {
		v.Set(defaultValue)
	}

	m.register(v)
	return v, nil
}

func (v *Var[T]) Key() string {
	return v.key
}

func (v *Var[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

func (v *Var[T]) Set(newValue T) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = newValue
	return v.syncNow()
}

func (v *Var[T]) SetLazy(newValue T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = newValue
	v.dirty = true
}

func (v *Var[T]) Sync() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.dirty {
		return nil
	}
	return v.syncNow()
}

func (v *Var[T]) syncNow() error {
	data, err := msgpack.Marshal(v.value)
	if err != nil {
		return err
	}

	// Optimización: Evita escribir en el almacenamiento si el valor no ha cambiado.
	if v.lastSynced != nil && bytes.Equal(v.lastSynced, data) {
		v.dirty = false
		return nil
	}

	if err := v.storage.Save(v.key, data); err != nil {
		return err
	}

	v.lastSynced = data
	v.dirty = false

	return nil
}