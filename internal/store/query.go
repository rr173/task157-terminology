package store

import (
	"context"
	"database/sql"
	"fmt"
)

// Delete removes one record. Services use it only for derived hit evidence;
// source documents, tasks and reviews are immutable audit subjects.
func (s *Store) Delete(ctx context.Context, kind, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM records WHERE kind=? AND id=?`, kind, id)
	if err != nil {
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) Count(ctx context.Context, kind, parent string) (int, error) {
	var count int
	var err error
	if parent == "" {
		err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM records WHERE kind=?`, kind).Scan(&count)
	} else {
		err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM records WHERE kind=? AND parent=?`, kind, parent).Scan(&count)
	}
	return count, err
}

func (s *Store) Require(ctx context.Context, kind, id string, v any) error {
	if err := s.Get(ctx, kind, id, v); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("%s %q not found: %w", kind, id, err)
		}
		return err
	}
	return nil
}

func (s *Store) ListAll(ctx context.Context, kind string, values any) error {
	return s.All(ctx, kind, values)
}
