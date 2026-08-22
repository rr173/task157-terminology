package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rr173/task157-terminology/internal/model"
	"github.com/rr173/task157-terminology/internal/store"
)

func TestReviewQueueFiltersAndPagesSuggestions(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	svc := New(db)
	library, err := svc.CreateLibrary(ctx, model.LibraryInput{Name: "review queue"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Publish(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	document, _, err := svc.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "queue-doc", Language: "zh", Fingerprint: "queue-fingerprint", Fragments: []model.FragmentInput{{Key: "first", Text: "存储 存储", Position: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Check(ctx, library.ID, []string{document.ID}); err != nil {
		t.Fatal(err)
	}
	queue, err := svc.ReviewQueue(ctx, model.ReviewQueueQuery{LibraryID: library.ID, Status: model.SuggestionOpen, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if queue.Total != 2 || len(queue.Items) != 1 || !queue.HasNextPage() || queue.OpenCount() != 2 {
		t.Fatalf("unexpected queue: %#v", queue)
	}
	pageTwo, err := svc.ReviewQueue(ctx, model.ReviewQueueQuery{LibraryID: library.ID, Language: "zh", Limit: 1, Offset: queue.NextOffset})
	if err != nil {
		t.Fatal(err)
	}
	if len(pageTwo.Items) != 1 || pageTwo.NextOffset != 0 {
		t.Fatalf("unexpected second page: %#v", pageTwo)
	}
}
