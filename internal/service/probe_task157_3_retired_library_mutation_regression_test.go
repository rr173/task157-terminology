package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"terminology/internal/httpapi"
	"terminology/internal/model"
	"terminology/internal/service"
	"terminology/internal/store"
)

func TestRetiredLibraryRejectsEveryMutationPath(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(filepath.Join(t.TempDir(), "retired-library.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := service.New(db)
	published := retiredLibrary(t, ctx, svc, "published library")
	if _, err = svc.Publish(ctx, published.ID); err == nil {
		t.Fatal("retired library was published again")
	}
	terms := retiredLibrary(t, ctx, svc, "terms library")
	if _, err = svc.AddTerm(ctx, terms.ID, model.TermInput{Concept: "edit", Language: "zh", Preferred: "编辑"}); err == nil {
		t.Fatal("retired library accepted a new term")
	}
	renamed := retiredLibrary(t, ctx, svc, "rename library")
	server := httptest.NewServer(httpapi.New(svc).Handler())
	defer server.Close()
	request, err := http.NewRequest(http.MethodPatch, server.URL+"/v1/libraries/"+renamed.ID, strings.NewReader(`{"name":"new library name","actor":"editor"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("retired library rename status=%d", response.StatusCode)
	}
}

func retiredLibrary(t *testing.T, ctx context.Context, svc *service.Service, name string) model.Library {
	t.Helper()
	library, err := svc.CreateLibrary(ctx, model.LibraryInput{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.AddTerm(ctx, library.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存"}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Publish(ctx, library.ID); err != nil {
		t.Fatal(err)
	}
	library, err = svc.RetireLibrary(ctx, library.ID, "owner")
	if err != nil {
		t.Fatal(err)
	}
	return library
}
