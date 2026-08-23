package service

import (
	"context"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/store"
)

func TestCurrentDocumentViewsIgnoreSupersededVersion(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "current-views.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	svc := New(db)
	library, err := svc.CreateLibrary(ctx, model.LibraryInput{Name: "current views"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Publish(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "old", Fragments: []model.FragmentInput{{Key: "intro", Text: "请存储文件", Position: 1}}}); err != nil {
		t.Fatal(err)
	}
	current, _, err := svc.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "new", Fragments: []model.FragmentInput{{Key: "intro", Text: "请保存文件", Position: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	documents, err := svc.CurrentDocuments(ctx, library.ID, "zh")
	if err != nil || len(documents) != 1 || documents[0].ID != current.ID {
		t.Fatalf("current documents: %v %#v", err, documents)
	}
	search, err := svc.SearchDocuments(ctx, model.DocumentQuery{LibraryID: library.ID, Language: "zh", Text: "存储"})
	if err != nil || len(search) != 0 {
		t.Fatalf("old wording must not be searchable: %v %#v", err, search)
	}
	coverage, err := svc.Coverage(ctx, library.ID)
	if err != nil || len(coverage.Languages) != 1 || coverage.Languages[0].Problematic != 0 {
		t.Fatalf("coverage must use current text: %v %#v", err, coverage)
	}
	usage, err := svc.TermUsage(ctx, library.ID)
	if err != nil || len(usage) != 1 || usage[0].ForbiddenHits != 0 || usage[0].Occurrences != 1 {
		t.Fatalf("usage must use current text: %v %#v", err, usage)
	}
}
