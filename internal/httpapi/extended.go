package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"terminology/internal/model"
)

func (a *API) libraryView(w http.ResponseWriter, r *http.Request) {
	library, terms, err := a.S.LibrarySnapshot(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, map[string]any{"library": library, "terms": terms})
}

func (a *API) libraryRename(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Actor string `json:"actor"`
	}
	if !decode(w, r, &body) {
		return
	}
	v, err := a.S.RenameLibrary(r.Context(), r.PathValue("id"), body.Name, body.Actor)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) libraryRetire(w http.ResponseWriter, r *http.Request) {
	actor := r.URL.Query().Get("actor")
	v, err := a.S.RetireLibrary(r.Context(), r.PathValue("id"), actor)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) libraryTerms(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.Terms(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) libraryDocuments(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	v, err := a.S.CurrentDocuments(r.Context(), r.PathValue("id"), language)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) batch(w http.ResponseWriter, r *http.Request) {
	var input model.ImportBatchInput
	if !decode(w, r, &input) {
		return
	}
	v, err := a.S.ImportBatch(r.Context(), r.PathValue("id"), input)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) batches(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.Batches(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) coverage(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.Coverage(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) export(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.ExportSnapshot(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) audit(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.AuditTrail(r.Context(), r.PathValue("id"), r.URL.Query().Get("entity_type"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) reviewQueue(w http.ResponseWriter, r *http.Request) {
	limit, err := optionalInt(r.URL.Query().Get("limit"))
	if err != nil {
		fail(w, err)
		return
	}
	offset, err := optionalInt(r.URL.Query().Get("offset"))
	if err != nil {
		fail(w, err)
		return
	}
	query := model.ReviewQueueQuery{LibraryID: r.PathValue("id"), Language: r.URL.Query().Get("language"), Concept: r.URL.Query().Get("concept"), Status: model.SuggestionStatus(r.URL.Query().Get("status")), Limit: limit, Offset: offset}
	v, err := a.S.ReviewQueue(r.Context(), query)
	if err != nil {
		write(w, v)
		return
	}
	write(w, v)
}

func optionalInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func (a *API) search(w http.ResponseWriter, r *http.Request) {
	query := model.DocumentQuery{LibraryID: r.PathValue("id"), Language: r.URL.Query().Get("language"), External: r.URL.Query().Get("external_id"), Status: r.URL.Query().Get("status"), Text: r.URL.Query().Get("q")}
	v, err := a.S.SearchDocuments(r.Context(), query)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) taskReport(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.TaskReport(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) diff(w http.ResponseWriter, r *http.Request) {
	v, err := a.S.Diff(r.Context(), r.URL.Query().Get("earlier"), r.URL.Query().Get("later"))
	if err != nil {
		fail(w, err)
		return
	}
	write(w, v)
}

func (a *API) review(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Accept bool   `json:"accept"`
		Actor  string `json:"actor"`
		Reason string `json:"reason"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	v, decision, err := a.S.ReviewWithActor(r.Context(), r.PathValue("id"), body.Accept, body.Actor, body.Reason)
	if err != nil {
		fail(w, err)
		return
	}
	write(w, map[string]any{"suggestion": v, "decision": decision})
}
