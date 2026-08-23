package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"terminology/internal/matcher"
	"terminology/internal/model"
)

func (s *Service) RenameLibrary(ctx context.Context, id, name, actor string) (model.Library, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(name) == "" {
		return model.Library{}, fmt.Errorf("library name is required")
	}
	var library model.Library
	if err := s.store.Require(ctx, "library", id, &library); err != nil {
		return library, err
	}
	if library.Status == model.LibraryRetired {
		return library, fmt.Errorf("retired library is immutable")
	}
	old := library.Name
	library.Name = strings.TrimSpace(name)
	library.Version++
	if err := s.store.Save(ctx, "library", id, "", library.Version, library); err != nil {
		return library, err
	}
	if err := s.recordAudit(ctx, id, "suggestion", id, "rename", actor, "library name changed", map[string]any{"from": old, "to": library.Name}); err != nil {
		return library, err
	}
	return library, nil
}

func (s *Service) RetireLibrary(ctx context.Context, id, actor string) (model.Library, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var library model.Library
	if err := s.store.Require(ctx, "library", id, &library); err != nil {
		return library, err
	}
	if library.Status == model.LibraryRetired {
		return library, nil
	}
	library.Status = model.LibraryRetired
	library.Version++
	if err := s.store.Save(ctx, "library", id, "", library.Version, library); err != nil {
		return library, err
	}
	if err := s.recordAudit(ctx, id, "library", id, "retire", actor, "library retired", nil); err != nil {
		return library, err
	}
	return library, nil
}

func (s *Service) LibrarySnapshot(ctx context.Context, id string) (model.Library, []model.Term, error) {
	library, err := s.Library(ctx, id)
	if err != nil {
		return library, nil, err
	}
	terms, err := s.Terms(ctx, id)
	if err != nil {
		return library, nil, err
	}
	sort.Slice(terms, func(i, j int) bool {
		if terms[i].Concept == terms[j].Concept {
			return terms[i].Language < terms[j].Language
		}
		return terms[i].Concept < terms[j].Concept
	})
	return library, terms, nil
}

func (s *Service) ValidateTermSet(ctx context.Context, libraryID string) ([]string, error) {
	terms, err := s.Terms(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	issues := make([]string, 0)
	seen := map[string]model.Term{}
	for _, term := range terms {
		key := strings.ToLower(term.Concept + "|" + term.Language)
		if previous, ok := seen[key]; ok && previous.Preferred != term.Preferred {
			issues = append(issues, fmt.Sprintf("%s has conflicting translations for %s", term.Concept, term.Language))
		}
		seen[key] = term
		if matcher.Normalize(term.Preferred) == "" {
			issues = append(issues, fmt.Sprintf("%s has an empty preferred translation", term.Concept))
		}
	}
	sort.Strings(issues)
	return issues, nil
}

func (s *Service) TermUsage(ctx context.Context, libraryID string) ([]model.TermUsage, error) {
	terms, err := s.Terms(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	usage := map[string]*model.TermUsage{}
	for _, term := range terms {
		key := term.Concept + "|" + term.Language
		usage[key] = &model.TermUsage{Concept: term.Concept, Language: term.Language, Preferred: term.Preferred}
	}
	for _, document := range documents {
		if !document.Status.Current() {
			continue
		}
		var fragments []model.Fragment
		if err := s.store.List(ctx, "fragment", document.ID, &fragments); err != nil {
			return nil, err
		}
		for _, fragment := range fragments {
			for _, hit := range matcher.Scan([]model.Fragment{fragment}, terms) {
				key := hit.Concept + "|" + hit.Language
				item := usage[key]
				if item == nil {
					continue
				}
				item.Occurrences++
				if hit.Forbidden {
					item.ForbiddenHits++
				}
			}
		}
	}
	result := make([]model.TermUsage, 0, len(usage))
	for _, item := range usage {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Concept < result[j].Concept })
	return result, nil
}
