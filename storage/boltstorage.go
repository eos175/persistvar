package storage

import (
	"errors"
	"slices"

	bolt "go.etcd.io/bbolt"
)

var defaultBucket = []byte("vars")

type BoltStorage struct {
	db *bolt.DB
}

func NewBoltStorage(path string) (*BoltStorage, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	// Crear bucket si no existe
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltStorage{
		db: db,
	}, nil
}

func (b *BoltStorage) Save(key string, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(defaultBucket).Put([]byte(key), value)
	})
}

func (b *BoltStorage) Load(key string) ([]byte, error) {
	var val []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(defaultBucket).Get([]byte(key))
		if v == nil {
			return errors.New("key not found")
		}
		val = slices.Clone(v) // copiar bytes
		return nil
	})
	return val, err
}

func (s *BoltStorage) Close() error {
	return s.db.Close()
}
