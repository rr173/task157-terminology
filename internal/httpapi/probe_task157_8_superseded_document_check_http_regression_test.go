package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/service"
	"terminology/internal/store"
)

func TestCheckRejectsSupersededDocument(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "superseded-check.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	ctx := context.Background(); core := service.New(db)
	library, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "superseded checks"}); if err != nil { t.Fatal(err) }
	if _, err = core.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存"}); err != nil { t.Fatal(err) }
	if _, err = core.Publish(ctx, library.ID); err != nil { t.Fatal(err) }
	old, _, err := core.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "old", Fragments: []model.FragmentInput{{Key: "p", Text: "保存", Position: 1}}}); if err != nil { t.Fatal(err) }
	if _, _, err = core.ImportDocument(ctx, library.ID, model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "new", Fragments: []model.FragmentInput{{Key: "p", Text: "保存", Position: 1}}}); err != nil { t.Fatal(err) }
	server := httptest.NewServer(New(core).Handler()); defer server.Close()
	body, _ := json.Marshal([]string{old.ID})
	response, err := http.Post(server.URL+"/v1/libraries/"+library.ID+"/check", "application/json", bytes.NewReader(body)); if err != nil { t.Fatal(err) }; defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest { t.Fatalf("superseded check status=%d", response.StatusCode) }
}
