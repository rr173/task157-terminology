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

func TestDraftLibraryRejectsSingleAndBatchImports(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "draft-import.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	core := service.New(db)
	library, err := core.CreateLibrary(context.Background(), model.LibraryInput{Name: "draft imports"})
	if err != nil { t.Fatal(err) }
	server := httptest.NewServer(New(core).Handler())
	defer server.Close()
	document := model.DocumentInput{ExternalID: "guide", Language: "zh", Fingerprint: "one", Fragments: []model.FragmentInput{{Key: "intro", Text: "请保存", Position: 1}}}
	body, err := json.Marshal(document)
	if err != nil { t.Fatal(err) }
	response, err := http.Post(server.URL+"/v1/libraries/"+library.ID+"/documents", "application/json", bytes.NewReader(body))
	if err != nil { t.Fatal(err) }
	if response.StatusCode != http.StatusBadRequest { t.Fatalf("single import status=%d", response.StatusCode) }
	response.Body.Close()
	batch, err := json.Marshal(model.ImportBatchInput{Source: "editor", Documents: []model.DocumentInput{document}})
	if err != nil { t.Fatal(err) }
	response, err = http.Post(server.URL+"/v1/libraries/"+library.ID+"/documents/batch", "application/json", bytes.NewReader(batch))
	if err != nil { t.Fatal(err) }
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest { t.Fatalf("batch import status=%d", response.StatusCode) }
}
