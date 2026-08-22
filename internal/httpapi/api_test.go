package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rr173/task157-terminology/internal/model"
	"github.com/rr173/task157-terminology/internal/service"
	"github.com/rr173/task157-terminology/internal/store"
)

func TestLibraryViewAndCoverageEndpoints(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	core := service.New(db)
	ctx := context.Background()
	lib, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "api"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = core.AddTerm(ctx, lib.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = core.Publish(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(core).Handler())
	defer server.Close()
	resp, err := http.Get(server.URL + "/v1/libraries/" + lib.ID)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("library endpoint failed: %v status=%d", err, resp.StatusCode)
	}
	resp.Body.Close()
	body, _ := json.Marshal(model.ImportBatchInput{Source: "api-test", Documents: []model.DocumentInput{{ExternalID: "d", Language: "zh", Fingerprint: "api-fp", Fragments: []model.FragmentInput{{Key: "p", Text: "请存储", Position: 0}}}}})
	resp, err = http.Post(server.URL+"/v1/libraries/"+lib.ID+"/documents/batch", "application/json", bytes.NewReader(body))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("batch endpoint failed: %v status=%d", err, resp.StatusCode)
	}
	resp.Body.Close()
	resp, err = http.Get(server.URL + "/v1/libraries/" + lib.ID + "/coverage")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("coverage endpoint failed: %v status=%d", err, resp.StatusCode)
	}
	resp.Body.Close()
}
