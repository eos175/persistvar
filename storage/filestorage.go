package storage

import (
	"os"
	"path/filepath"
)

type FileStorage struct {
	dir string
}

func NewFileStorage(dir string) (*FileStorage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{dir: dir}, nil
}

func (fs *FileStorage) Save(key string, value []byte) error {
	path := filepath.Join(fs.dir, key+".var")
	return os.WriteFile(path, value, 0644)
}

func (fs *FileStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(fs.dir, key+".var")
	return os.ReadFile(path)
}

func (fs *FileStorage) Close() error {
	return nil
}
