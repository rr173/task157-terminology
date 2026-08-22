package service

import (
	"context"
	"path/filepath"
	"terminology/internal/model"
	"terminology/internal/store"
	"testing"
)

func TestWorkflow(t *testing.T) {
	db, e := store.Open(filepath.Join(t.TempDir(), "x.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := New(db)
	ctx := context.Background()
	l, e := s.CreateLibrary(ctx, model.LibraryInput{Name: "docs"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.AddTerm(ctx, l.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Publish(ctx, l.ID); e != nil {
		t.Fatal(e)
	}
	d, _, e := s.ImportDocument(ctx, l.ID, model.DocumentInput{ExternalID: "a", Language: "zh", Fingerprint: "f", Fragments: []model.FragmentInput{{Key: "1", Text: "请存储", Position: 1}}})
	if e != nil {
		t.Fatal(e)
	}
	task, items, e := s.Check(ctx, l.ID, []string{d.ID})
	if e != nil || task.Status != model.TaskSucceeded || len(items) == 0 {
		t.Fatal(task, items, e)
	}
	if _, replayed, e := s.ImportDocument(ctx, l.ID, model.DocumentInput{ExternalID: "a", Language: "zh", Fingerprint: "replacement", Fragments: []model.FragmentInput{{Key: "1", Text: "请保存", Position: 1}}}); e != nil || replayed {
		t.Fatalf("replace document: replayed=%v err=%v", replayed, e)
	}
	updated, e := s.Suggestions(ctx, task.ID)
	if e != nil || updated[0].Status != model.SuggestionExpired {
		t.Fatalf("old suggestion should expire after replacement: %#v err=%v", updated, e)
	}
}
