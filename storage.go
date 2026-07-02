package persistvar

import (
	"encoding/json"
	"errors"
)

var ErrNotFound = errors.New("persistvar: key not found")

// Storage defines the interface for persistent storage backends.
// Implementations must ensure thread-safety where applicable for their storage mechanisms.
type Storage interface {
	// Save writes a new value for a given key.
	// It includes the oldValue for potential optimistic locking or Compare-And-Swap (CAS) like operations,
	// allowing implementations to prevent overwrites if the value has changed since it was last read.
	// Implementations may use oldValue to optimize writes by comparing it with newValue; if they are equal, the operation can be skipped.
	Save(key string, newValue []byte, oldValue []byte) error
	// Load retrieves the value associated with a key.
	// Returns an error if the key does not exist.
	Load(key string) ([]byte, error)
	// Delete removes a key-value pair from the storage.
	Delete(key string) error
	// Close releases any resources held by the storage backend.
	Close() error
}

// Serializer defines the interface for variable serialization.
type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// JSONSerializer is a default implementation using encoding/json.
type JSONSerializer struct{}

func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
