package persistvar

import (
	"context"
	"errors"
	"sync"
	"time"
)

// VarManager manages a collection of persistent variables, providing lifecycle
// and synchronization mechanisms.
type VarManager struct {
	storage    Storage
	serializer Serializer
	vars       []syncable
	registry   map[string]syncable
	mu         sync.Mutex
	autosyncCh chan struct{}
}

type syncable interface {
	Sync() error
}

// NewVarManager creates a new manager for persistent variables, using the provided storage backend.
// If serializer is nil, JSONSerializer is used by default.
func NewVarManager(storage Storage, serializer Serializer) *VarManager {
	if serializer == nil {
		serializer = &JSONSerializer{}
	}
	return &VarManager{
		storage:    storage,
		serializer: serializer,
		registry:   make(map[string]syncable),
	}
}

// loadOrStore returns the existing variable if already registered,
// or creates a new one using the factory function, registers it, and returns it.
func (m *VarManager) loadOrStore(key string, factory func() (syncable, error)) (syncable, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.registry[key]; ok {
		return existing, nil
	}

	v, err := factory()
	if err != nil {
		return nil, err
	}

	m.vars = append(m.vars, v)
	m.registry[key] = v
	return v, nil
}

// Sync forces all managed variables to write their pending changes to storage immediately.
// It returns an aggregate of all errors encountered during synchronization.
func (m *VarManager) Sync() error {
	m.mu.Lock()
	vars := append([]syncable(nil), m.vars...)
	m.mu.Unlock()

	var errs []error
	for _, v := range vars {
		if err := v.Sync(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// AutoSync starts a goroutine that periodically saves all lazy variables to storage.
// It returns immediately if AutoSync is already running.
func (m *VarManager) AutoSync(ctx context.Context, interval time.Duration) {
	if m.autosyncCh != nil {
		return // autosync is already running
	}

	m.autosyncCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.autosyncCh:
				return
			case <-ticker.C:
				m.Sync()
			}
		}
	}()
}

// StopAutoSync halts the automatic synchronization process.
func (m *VarManager) StopAutoSync() {
	if m.autosyncCh != nil {
		close(m.autosyncCh)
		m.autosyncCh = nil
	}
}

// Close gracefully shuts down the manager, ensuring all pending changes are written to storage
// and stopping any active AutoSync routine.
func (m *VarManager) Close() error {
	if err := m.Sync(); err != nil {
		return err
	}
	m.StopAutoSync()
	return m.storage.Close()
}
