package persistvar

type Storage interface {
	Save(key string, newValue []byte, oldValue []byte) error
	Load(key string) ([]byte, error)
	Delete(key string) error
	Close() error
}
