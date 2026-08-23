package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/service"
	"terminology/internal/store"
)

func TestDocumentDiffReportsExpiredSuggestions(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "diff-invalidations.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	ctx := context.Background(); core := service.New(db)
	library, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "diff invalidations"}); if err != nil { t.Fatal(err) }
	if _, err = core.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); err != nil { t.Fatal(err) }
	if _, err = core.Publish(ctx, library.ID); err != nil { t.Fatal(err) }
	old, _, err := core.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "old", Fragments: []model.FragmentInput{{Key: "intro", Text: "请存储", Position: 1}}}); if err != nil { t.Fatal(err) }
	if _, _, err = core.Check(ctx, library.ID, []string{old.ID}); err != nil { t.Fatal(err) }
	newDoc, _, err := core.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "new", Fragments: []model.FragmentInput{{Key: "intro", Text: "请保存", Position: 1}}}); if err != nil { t.Fatal(err) }
	server := httptest.NewServer(New(core).Handler()); defer server.Close()
	query := url.Values{"earlier": {old.ID}, "later": {newDoc.ID}}
	response, err := http.Get(server.URL + "/v1/documents/diff?" + query.Encode()); if err != nil { t.Fatal(err) }; defer response.Body.Close()
	if response.StatusCode != http.StatusOK { t.Fatalf("diff status=%d", response.StatusCode) }
	var diff model.Diff; if err = json.NewDecoder(response.Body).Decode(&diff); err != nil { t.Fatal(err) }
	if diff.InvalidatedSuggestions != 1 { t.Fatalf("invalidated=%d", diff.InvalidatedSuggestions) }
}
