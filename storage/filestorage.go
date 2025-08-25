package storage

import (
	"bytes"
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

func (fs *FileStorage) Save(key string, newValue []byte, oldValue []byte) error {
	// Optimización: si el valor no ha cambiado, no hacemos nada.
	if oldValue != nil && bytes.Equal(newValue, oldValue) {
		return nil
	}

	path := filepath.Join(fs.dir, key+".var")
	tmpPath := path + ".tmp"

	// Escribir primero a un archivo temporal
	if err := os.WriteFile(tmpPath, newValue, 0644); err != nil {
		return err
	}

	// Reemplazo atómico
	return os.Rename(tmpPath, path)
}

func (fs *FileStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(fs.dir, key+".var")
	return os.ReadFile(path)
}

func (fs *FileStorage) Delete(key string) error {
	path := filepath.Join(fs.dir, key+".var")
	// os.Remove devuelve un error si el archivo no existe,
	// lo cual está bien para nosotros.
	return os.Remove(path)
}

func (fs *FileStorage) Close() error {
	return nil
}
