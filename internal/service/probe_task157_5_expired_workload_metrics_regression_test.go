package service

import (
	"context"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/store"
)

func TestExpiredSuggestionsLeaveNoOpenWorkloadMetrics(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "expired-metrics.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	svc := New(db)
	library, err := svc.CreateLibrary(ctx, model.LibraryInput{Name: "expired metrics"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Publish(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	old, _, err := svc.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "old", Fragments: []model.FragmentInput{{Key: "intro", Text: "请存储文件", Position: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.Check(ctx, library.ID, []string{old.ID}); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "new", Fragments: []model.FragmentInput{{Key: "intro", Text: "请保存文件", Position: 1}}}); err != nil {
		t.Fatal(err)
	}
	report, err := svc.LibraryReport(ctx, library.ID)
	if err != nil || report.OpenSuggestions != 0 {
		t.Fatalf("library report still has open work: %v %#v", err, report)
	}
	health, err := svc.HealthReport(ctx)
	if err != nil || health.OpenReviews != 0 {
		t.Fatalf("health still has open reviews: %v %#v", err, health)
	}
	coverage, err := svc.Coverage(ctx, library.ID)
	if err != nil || coverage.Suggestions != 0 {
		t.Fatalf("coverage still has open suggestions: %v %#v", err, coverage)
	}
}
