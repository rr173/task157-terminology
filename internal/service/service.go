package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"terminology/internal/consistency"
	"terminology/internal/matcher"
	"terminology/internal/model"
	"terminology/internal/store"
	"time"
)

type Service struct {
	store *store.Store
	mu    sync.Mutex
	now   func() time.Time
}

func New(store *store.Store) *Service { return &Service{store: store, now: time.Now} }
func (s *Service) CreateLibrary(ctx context.Context, input model.LibraryInput) (model.Library, error) {
	if err := input.Validate(); err != nil {
		return model.Library{}, err
	}
	v := model.Library{ID: uuid.NewString(), Name: input.Name, Version: 1, Status: model.LibraryDraft, CreatedAt: s.now().UTC()}
	return v, s.store.Save(ctx, "library", v.ID, "", v.Version, v)
}
func (s *Service) Library(ctx context.Context, id string) (model.Library, error) {
	var v model.Library
	return v, s.store.Get(ctx, "library", id, &v)
}
func (s *Service) AddTerm(ctx context.Context, libraryID string, input model.TermInput) (model.Term, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := input.Validate(); err != nil {
		return model.Term{}, err
	}
	library, err := s.Library(ctx, libraryID)
	if err != nil {
		return model.Term{}, err
	}
	if library.Status == model.LibraryRetired {
		return model.Term{}, fmt.Errorf("retired library is immutable")
	}
	v := model.Term{ID: uuid.NewString(), LibraryID: libraryID, Concept: input.Concept, Language: input.Language, Preferred: input.Preferred, Forbidden: input.Forbidden, CreatedAt: s.now().UTC()}
	return v, s.store.Save(ctx, "term", v.ID, libraryID, library.Version, v)
}
func (s *Service) Terms(ctx context.Context, libraryID string) ([]model.Term, error) {
	var v []model.Term
	return v, s.store.List(ctx, "term", libraryID, &v)
}
func (s *Service) Publish(ctx context.Context, libraryID string) (model.Library, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, err := s.Library(ctx, libraryID)
	if err != nil {
		return v, err
	}
	if v.Status == model.LibraryRetired {
		return v, fmt.Errorf("retired library is immutable")
	}
	terms, err := s.Terms(ctx, libraryID)
	if err != nil {
		return v, err
	}
	if len(terms) == 0 {
		return v, fmt.Errorf("library requires terms")
	}
	issues, err := s.ValidateTermSet(ctx, libraryID)
	if err != nil {
		return v, err
	}
	if len(issues) > 0 {
		return v, fmt.Errorf("library has conflicting terms: %s", issues[0])
	}
	v.Status = model.LibraryPublished
	v.Version++
	return v, s.store.Save(ctx, "library", v.ID, "", v.Version, v)
}
func (s *Service) ImportDocument(ctx context.Context, libraryID string, input model.DocumentInput) (model.Document, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := input.Validate(); err != nil {
		return model.Document{}, false, err
	}
	library, err := s.Library(ctx, libraryID)
	if err != nil {
		return model.Document{}, false, err
	}
	if library.Status != model.LibraryPublished {
		return model.Document{}, false, fmt.Errorf("library must be published")
	}
	var docs []model.Document
	if err = s.store.List(ctx, "document", libraryID, &docs); err != nil {
		return model.Document{}, false, err
	}
	for _, doc := range docs {
		if doc.Fingerprint == input.Fingerprint {
			return doc, true, nil
		}
	}
	version := 1
	for _, doc := range docs {
		if doc.ExternalID == input.ExternalID && doc.Version >= version {
			version = doc.Version + 1
			doc.Status = model.DocumentSuperseded
			if err = s.store.Save(ctx, "document", doc.ID, libraryID, doc.Version, doc); err != nil {
				return model.Document{}, false, err
			}
			if _, err = s.expireOpenSuggestionsForDocument(ctx, doc.ID); err != nil {
				return model.Document{}, false, err
			}
		}
	}
	v := model.Document{ID: uuid.NewString(), LibraryID: libraryID, Fingerprint: input.Fingerprint, ExternalID: input.ExternalID, Language: input.Language, Version: version, Status: model.DocumentImported, CreatedAt: s.now().UTC()}
	if err = s.store.Save(ctx, "document", v.ID, libraryID, v.Version, v); err != nil {
		return v, false, err
	}
	for _, item := range input.Fragments {
		fragment := model.Fragment{ID: uuid.NewString(), DocumentID: v.ID, Key: item.Key, Text: item.Text, Position: item.Position, CreatedAt: s.now().UTC()}
		if err = s.store.Save(ctx, "fragment", fragment.ID, v.ID, item.Position, fragment); err != nil {
			return v, false, err
		}
	}
	return v, false, nil
}
func (s *Service) Documents(ctx context.Context, libraryID string) ([]model.Document, error) {
	var v []model.Document
	return v, s.store.List(ctx, "document", libraryID, &v)
}
func (s *Service) Check(ctx context.Context, libraryID string, ids []string) (model.CheckTask, []model.Suggestion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	library, err := s.Library(ctx, libraryID)
	if err != nil {
		return model.CheckTask{}, nil, err
	}
	if library.Status != model.LibraryPublished {
		return model.CheckTask{}, nil, fmt.Errorf("library is not published")
	}
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return model.CheckTask{}, nil, err
	}
	selected := []model.Document{}
	for _, doc := range documents {
		for _, id := range ids {
			if doc.ID == id && doc.Status != model.DocumentSuperseded {
				selected = append(selected, doc)
			}
		}
	}
	if len(selected) == 0 {
		return model.CheckTask{}, nil, fmt.Errorf("no current documents selected")
	}
	task := model.CheckTask{ID: uuid.NewString(), LibraryID: libraryID, Status: model.TaskRunning, CreatedAt: s.now().UTC()}
	for _, doc := range selected {
		task.DocumentIDs = append(task.DocumentIDs, doc.ID)
	}
	task.Snapshot = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(task.DocumentIDs))))
	_ = s.store.Save(ctx, "task", task.ID, libraryID, 0, task)
	terms, err := s.Terms(ctx, libraryID)
	if err != nil {
		return task, nil, err
	}
	fragments := []model.Fragment{}
	hits := []model.Hit{}
	for _, doc := range selected {
		var current []model.Fragment
		if err = s.store.List(ctx, "fragment", doc.ID, &current); err != nil {
			return task, nil, err
		}
		fragments = append(fragments, current...)
		for _, fragment := range current {
			for _, hit := range matcher.Hits(fragment, terms) {
				hit.ID = uuid.NewString()
				hits = append(hits, hit)
				if err = s.store.Save(ctx, "hit", hit.ID, fragment.ID, hit.Start, hit); err != nil {
					return task, nil, err
				}
			}
		}
	}
	suggestions := consistency.Suggestions(task, selected, fragments, hits, terms)
	for i := range suggestions {
		suggestions[i].ID = uuid.NewString()
		if err = s.store.Save(ctx, "suggestion", suggestions[i].ID, task.ID, 0, suggestions[i]); err != nil {
			return task, nil, err
		}
	}
	task.Status = model.TaskSucceeded
	task.FinishedAt = s.now().UTC()
	if err = s.store.Save(ctx, "task", task.ID, libraryID, 0, task); err != nil {
		return task, nil, err
	}
	for _, doc := range selected {
		doc.Status = model.DocumentChecked
		_ = s.store.Save(ctx, "document", doc.ID, libraryID, doc.Version, doc)
	}
	return task, suggestions, nil
}
