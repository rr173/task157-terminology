package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestCountRequireAndDelete(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "query.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if err = s.Save(ctx, "x", "1", "p", 1, map[string]string{"value": "ok"}); err != nil {
		t.Fatal(err)
	}
	if count, err := s.Count(ctx, "x", "p"); err != nil || count != 1 {
		t.Fatalf("count mismatch: %d %v", count, err)
	}
	var value map[string]string
	if err = s.Require(ctx, "x", "1", &value); err != nil || value["value"] != "ok" {
		t.Fatalf("require failed: %#v %v", value, err)
	}
	if err = s.Delete(ctx, "x", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Count(ctx, "x", "p"); err != nil {
		t.Fatal(err)
	}
}
