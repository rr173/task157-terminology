package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"terminology/internal/service"
	"terminology/internal/store"
)

func TestReviewQueueRejectsUnknownStatus(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := httptest.NewServer(New(service.New(db)).Handler())
	defer server.Close()
	response, err := http.Get(server.URL + "/v1/libraries/any/review-queue?status=stale")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown status response=%d", response.StatusCode)
	}
}
