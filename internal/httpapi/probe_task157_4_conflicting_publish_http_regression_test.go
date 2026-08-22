package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/service"
	"terminology/internal/store"
)

func TestPublishRejectsConflictingTranslations(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "conflicting-terms.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	core := service.New(db)
	library, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "conflicting translations"})
	if err != nil {
		t.Fatal(err)
	}
	for _, preferred := range []string{"保存", "存盘"} {
		if _, err = core.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: preferred}); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(New(core).Handler())
	defer server.Close()
	response, err := http.Post(server.URL+"/v1/libraries/"+library.ID+"/publish", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("conflicting publish status=%d", response.StatusCode)
	}
	stored, err := core.Library(ctx, library.ID)
	if err != nil || stored.Status != model.LibraryDraft || stored.Version != 1 {
		t.Fatalf("conflicting library was published: %v %#v", err, stored)
	}
}
