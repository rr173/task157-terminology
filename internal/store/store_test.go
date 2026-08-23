package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoreHealth(t *testing.T) {
	s, e := Open(filepath.Join(t.TempDir(), "x.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if e = s.Ping(context.Background()); e != nil {
		t.Fatal(e)
	}
}
