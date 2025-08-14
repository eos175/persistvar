package storage

import (
	"database/sql"

	_ "modernc.org/sqlite" // driver SQLite puro Go
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Crear tabla si no existe
	query := `
    CREATE TABLE IF NOT EXISTS vars (
        key TEXT PRIMARY KEY,
        value BLOB
    );`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Save(key string, value []byte) error {
	_, err := s.db.Exec(`
        INSERT INTO vars(key, value) VALUES (?, ?)
        ON CONFLICT(key) DO UPDATE SET value=excluded.value
    `, key, value)
	return err
}

func (s *SQLiteStorage) Load(key string) ([]byte, error) {
	row := s.db.QueryRow("SELECT value FROM vars WHERE key=?", key)
	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
