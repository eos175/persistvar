package persistvar

import (
	"encoding/json"
	"sync"
)

type Var[T any] struct {
	key     string
	value   T
	storage Storage
	mu      sync.RWMutex
	dirty   bool
}

func NewVar[T any](m *VarManager, key string, defaultValue T) (*Var[T], error) {
	v := &Var[T]{key: key, storage: m.storage}

	data, err := m.storage.Load(key)
	if err == nil {
		if unmarshalErr := json.Unmarshal(data, &v.value); unmarshalErr != nil {
			return nil, unmarshalErr
		}
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
	data, err := json.Marshal(v.value)
	if err != nil {
		return err
	}
	v.dirty = false
	return v.storage.Save(v.key, data)
}
