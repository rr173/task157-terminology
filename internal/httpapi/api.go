package httpapi

import (
	"encoding/json"
	"net/http"
	"terminology/internal/model"
	"terminology/internal/service"
)

type API struct{ S *service.Service }

func New(s *service.Service) *API { return &API{s} }
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("POST /v1/libraries", a.library)
	m.HandleFunc("GET /v1/libraries/{id}", a.libraryView)
	m.HandleFunc("PATCH /v1/libraries/{id}", a.libraryRename)
	m.HandleFunc("POST /v1/libraries/{id}/retire", a.libraryRetire)
	m.HandleFunc("GET /v1/libraries/{id}/terms", a.libraryTerms)
	m.HandleFunc("GET /v1/libraries/{id}/documents", a.libraryDocuments)
	m.HandleFunc("GET /v1/libraries/{id}/documents/search", a.search)
	m.HandleFunc("GET /v1/libraries/{id}/coverage", a.coverage)
	m.HandleFunc("GET /v1/libraries/{id}/export", a.export)
	m.HandleFunc("GET /v1/libraries/{id}/audit", a.audit)
	m.HandleFunc("GET /v1/libraries/{id}/review-queue", a.reviewQueue)
	m.HandleFunc("GET /v1/libraries/{id}/batches", a.batches)
	m.HandleFunc("POST /v1/libraries/{id}/terms", a.term)
	m.HandleFunc("POST /v1/libraries/{id}/publish", a.publish)
	m.HandleFunc("POST /v1/libraries/{id}/documents", a.document)
	m.HandleFunc("POST /v1/libraries/{id}/documents/batch", a.batch)
	m.HandleFunc("POST /v1/libraries/{id}/check", a.check)
	m.HandleFunc("GET /v1/tasks/{id}/suggestions", a.suggestions)
	m.HandleFunc("GET /v1/tasks/{id}/report", a.taskReport)
	m.HandleFunc("POST /v1/suggestions/{id}/accept", a.accept)
	m.HandleFunc("POST /v1/suggestions/{id}/ignore", a.ignore)
	m.HandleFunc("POST /v1/suggestions/{id}/review", a.review)
	m.HandleFunc("GET /v1/documents/diff", a.diff)
	m.HandleFunc("POST /v1/recovery", a.recover)
	return m
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, err.Error(), 400)
		return false
	}
	return true
}
func write(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, e error) { http.Error(w, e.Error(), 400) }
func (a *API) library(w http.ResponseWriter, r *http.Request) {
	var in model.LibraryInput
	if !decode(w, r, &in) {
		return
	}
	v, e := a.S.CreateLibrary(r.Context(), in)
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) term(w http.ResponseWriter, r *http.Request) {
	var in model.TermInput
	if !decode(w, r, &in) {
		return
	}
	v, e := a.S.AddTerm(r.Context(), r.PathValue("id"), in)
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) publish(w http.ResponseWriter, r *http.Request) {
	v, e := a.S.Publish(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) document(w http.ResponseWriter, r *http.Request) {
	var in model.DocumentInput
	if !decode(w, r, &in) {
		return
	}
	v, _, e := a.S.ImportDocument(r.Context(), r.PathValue("id"), in)
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) check(w http.ResponseWriter, r *http.Request) {
	var ids []string
	if !decode(w, r, &ids) {
		return
	}
	v, x, e := a.S.Check(r.Context(), r.PathValue("id"), ids)
	if e != nil {
		write(w, map[string]any{"task": v, "suggestions": x})
		return
	}
	write(w, map[string]any{"task": v, "suggestions": x})
}
func (a *API) suggestions(w http.ResponseWriter, r *http.Request) {
	v, e := a.S.Suggestions(r.Context(), r.PathValue("id"))
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) accept(w http.ResponseWriter, r *http.Request) {
	v, e := a.S.Review(r.Context(), r.PathValue("id"), true)
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) ignore(w http.ResponseWriter, r *http.Request) {
	v, e := a.S.Review(r.Context(), r.PathValue("id"), false)
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
func (a *API) recover(w http.ResponseWriter, r *http.Request) {
	v, e := a.S.Recover(r.Context())
	if e != nil {
		fail(w, e)
		return
	}
	write(w, v)
}
