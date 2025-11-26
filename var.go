package persistvar

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// Var is a generic variable that persists its value to storage.
type Var[T any] struct {
	key        string
	value      T
	storage    Storage
	mu         sync.RWMutex
	dirty      bool
	lastSynced []byte // Snapshot of the last saved state
}

// NewVar creates or retrieves a managed persistent variable.
// If a Var with the given key already exists, the existing instance is returned.
// Otherwise, a new Var is initialized with defaultValue.
//
// Important: For the same key, NewVar always returns the identical *Var[T] instance.
func NewVar[T any](m *VarManager, key string, defaultValue T) (*Var[T], error) {
	factory := func() (syncable, error) {
		v := &Var[T]{key: key, storage: m.storage}

		data, err := m.storage.Load(key)
		if err == nil {
			if unmarshalErr := msgpack.Unmarshal(data, &v.value); unmarshalErr != nil {
				return nil, unmarshalErr
			}
			v.lastSynced = data
		} else {
			// Not found in disk, initialize with default and mark for saving later.
			v.SetLazy(defaultValue)
		}
		return v, nil
	}

	// Delegate everything to the manager
	obj, err := m.loadOrStore(key, factory)
	if err != nil {
		return nil, err
	}

	if typed, ok := obj.(*Var[T]); ok {
		return typed, nil
	}

	return nil, fmt.Errorf("var '%s' already exists with a different type", key)
}

// Key returns the unique identifier for this variable.
func (v *Var[T]) Key() string {
	return v.key
}

// Get returns the current value.
func (v *Var[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

// Set updates the value and immediately writes it to storage.
func (v *Var[T]) Set(newValue T) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = newValue
	return v.syncNow()
}

// SetLazy updates the value in memory only.
// The change will be written to storage during the next Sync or background save.
func (v *Var[T]) SetLazy(newValue T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = newValue
	v.dirty = true
}

// UpdateLazy atomically modifies the value in memory using the transform function.
// Returns the new value. If 'changed' is false, no update occurs.
func (v *Var[T]) UpdateLazy(transform func(T) (T, bool)) T {
	v.mu.Lock()
	defer v.mu.Unlock()

	newValue, changed := transform(v.value)
	if !changed {
		return v.value
	}

	v.value = newValue
	v.dirty = true
	return newValue
}

// Update atomically modifies the value and immediately writes to storage if changed.
// Returns the new value and any storage error.
func (v *Var[T]) Update(transform func(T) (T, bool)) (T, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	newValue, changed := transform(v.value)
	if !changed {
		return v.value, nil
	}

	v.value = newValue
	if err := v.syncNow(); err != nil {
		return newValue, err
	}

	return newValue, nil
}

// Sync forces a write to storage if there are pending changes.
func (v *Var[T]) Sync() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.dirty {
		return nil
	}
	return v.syncNow()
}

// syncNow performs the actual synchronization to storage. It expects the mutex to be locked.
func (v *Var[T]) syncNow() error {
	data, err := msgpack.Marshal(v.value)
	if err != nil {
		return err
	}

	if err := v.storage.Save(v.key, data, v.lastSynced); err != nil {
		return err
	}

	v.lastSynced = data
	v.dirty = false

	return nil
}
