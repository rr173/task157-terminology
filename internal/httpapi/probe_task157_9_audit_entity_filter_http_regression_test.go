package httpapi

import (
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

func TestAuditEndpointFiltersByEntityType(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "audit-filter.db")); if err != nil { t.Fatal(err) }; defer db.Close()
	core := service.New(db); ctx := context.Background()
	library, err := core.CreateLibrary(ctx, model.LibraryInput{Name: "audit filter"}); if err != nil { t.Fatal(err) }
	if _, err = core.RenameLibrary(ctx, library.ID, "renamed", "editor"); err != nil { t.Fatal(err) }
	server := httptest.NewServer(New(core).Handler()); defer server.Close()
	response, err := http.Get(server.URL+"/v1/libraries/"+library.ID+"/audit?entity_type=suggestion"); if err != nil { t.Fatal(err) }; defer response.Body.Close()
	if response.StatusCode != http.StatusOK { t.Fatalf("audit status=%d", response.StatusCode) }
	var records []model.AuditRecord; if err = json.NewDecoder(response.Body).Decode(&records); err != nil { t.Fatal(err) }
	if len(records) != 0 { t.Fatalf("unexpected filtered records=%#v", records) }
}
