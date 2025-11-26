package persistvar

// Storage defines the interface for persistent storage backends.
// Implementations must ensure thread-safety where applicable for their storage mechanisms.
type Storage interface {
	// Save writes a new value for a given key.
	// It includes the oldValue for potential optimistic locking or Compare-And-Swap (CAS) like operations,
	// allowing implementations to prevent overwrites if the value has changed since it was last read.
	Save(key string, newValue []byte, oldValue []byte) error
	// Load retrieves the value associated with a key.
	// Returns an error if the key does not exist.
	Load(key string) ([]byte, error)
	// Delete removes a key-value pair from the storage.
	Delete(key string) error
	// Close releases any resources held by the storage backend.
	Close() error
}
