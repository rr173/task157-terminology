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

func TestPartialBatchKeepsAcceptedDocumentAndReportsRejectedDocument(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "partial-batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	core := service.New(db)
	library, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "partial batches"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = core.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存"}); err != nil {
		t.Fatal(err)
	}
	if _, err = core.Publish(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(core).Handler())
	defer server.Close()
	input := model.ImportBatchInput{Source: "editor", Documents: []model.DocumentInput{
		{ExternalID: "valid", Language: "zh", Fingerprint: "valid-fingerprint", Fragments: []model.FragmentInput{{Key: "intro", Text: "请保存文件", Position: 1}}},
		{ExternalID: "invalid", Language: "zh", Fingerprint: "invalid-fingerprint"},
	}}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(server.URL+"/v1/libraries/"+library.ID+"/documents/batch", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("partial batch status=%d", response.StatusCode)
	}
	var batch model.ImportBatch
	if err = json.NewDecoder(response.Body).Decode(&batch); err != nil {
		t.Fatal(err)
	}
	if batch.Status != model.ImportBatchPartial || batch.Accepted != 1 || batch.Rejected != 1 || len(batch.DocumentIDs) != 1 || len(batch.Errors) != 1 {
		t.Fatalf("partial result was lost: %#v", batch)
	}
	documents, err := core.CurrentDocuments(ctx, library.ID, "zh")
	if err != nil || len(documents) != 1 || documents[0].ExternalID != "valid" {
		t.Fatalf("accepted document was not retained: %v %#v", err, documents)
	}
}
