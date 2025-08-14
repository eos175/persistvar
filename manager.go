package persistvar

import (
	"context"
	"sync"
	"time"
)

type VarManager struct {
	storage    Storage
	vars       []syncable
	mu         sync.Mutex
	autosyncCh chan struct{}
}

type syncable interface {
	Sync() error
}

func NewVarManager(storage Storage) *VarManager {
	return &VarManager{storage: storage}
}

func (m *VarManager) register(v syncable) {
	m.mu.Lock()
	m.vars = append(m.vars, v)
	m.mu.Unlock()
}

func (m *VarManager) Sync() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.vars {
		if err := v.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// AutoSync inicia un goroutine que guarda todas las variables lazy periódicamente
func (m *VarManager) AutoSync(ctx context.Context, interval time.Duration) {
	if m.autosyncCh != nil {
		return // ya hay un autosync en ejecución
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

// StopAutoSync permite detener el autosync manualmente
func (m *VarManager) StopAutoSync() {
	if m.autosyncCh != nil {
		close(m.autosyncCh)
		m.autosyncCh = nil
	}
}

func (m *VarManager) Close() error {
	if err := m.Sync(); err != nil {
		return err
	}
	m.StopAutoSync()
	return m.storage.Close()
}
