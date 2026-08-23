package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db}
	if err = s.init(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error                   { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) init(ctx context.Context) error {
	for _, q := range []string{`PRAGMA foreign_keys=ON`, `CREATE TABLE IF NOT EXISTS records(kind TEXT NOT NULL,id TEXT NOT NULL,parent TEXT NOT NULL,version INTEGER NOT NULL,body BLOB NOT NULL,PRIMARY KEY(kind,id))`, `CREATE INDEX IF NOT EXISTS records_kind_parent_idx ON records(kind,parent,version)`} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("schema: %w", err)
		}
	}
	return nil
}
func (s *Store) Save(ctx context.Context, kind, id, parent string, version int, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO records(kind,id,parent,version,body) VALUES(?,?,?,?,?) ON CONFLICT(kind,id) DO UPDATE SET parent=excluded.parent,version=excluded.version,body=excluded.body`, kind, id, parent, version, data)
	return err
}
func (s *Store) Get(ctx context.Context, kind, id string, v any) error {
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT body FROM records WHERE kind=? AND id=?`, kind, id).Scan(&data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func (s *Store) List(ctx context.Context, kind, parent string, values any) error {
	rows, err := s.db.QueryContext(ctx, `SELECT body FROM records WHERE kind=? AND parent=? ORDER BY version,id`, kind, parent)
	if err != nil {
		return err
	}
	defer rows.Close()
	raw := make([]json.RawMessage, 0)
	for rows.Next() {
		var body []byte
		if err = rows.Scan(&body); err != nil {
			return err
		}
		raw = append(raw, body)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, values)
}
func (s *Store) All(ctx context.Context, kind string, values any) error {
	rows, err := s.db.QueryContext(ctx, `SELECT body FROM records WHERE kind=? ORDER BY parent,version,id`, kind)
	if err != nil {
		return err
	}
	defer rows.Close()
	raw := []json.RawMessage{}
	for rows.Next() {
		var body []byte
		if err = rows.Scan(&body); err != nil {
			return err
		}
		raw = append(raw, body)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, values)
}
func IsNoRows(err error) bool { return err == sql.ErrNoRows }
