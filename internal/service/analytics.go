package service

import (
	"context"
	"sort"

	"github.com/rr173/task157-terminology/internal/matcher"
	"github.com/rr173/task157-terminology/internal/model"
)

func (s *Service) Coverage(ctx context.Context, libraryID string) (model.CoverageReport, error) {
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return model.CoverageReport{}, err
	}
	terms, err := s.Terms(ctx, libraryID)
	if err != nil {
		return model.CoverageReport{}, err
	}
	report := model.CoverageReport{LibraryID: libraryID, Documents: len(documents), GeneratedAt: s.now().UTC()}
	byLanguage := map[string]*model.LanguageCoverage{}
	for _, document := range documents {
		if document.Status == model.DocumentSuperseded {
			continue
		}
		coverage := byLanguage[document.Language]
		if coverage == nil {
			coverage = &model.LanguageCoverage{Language: document.Language}
			byLanguage[document.Language] = coverage
		}
		coverage.Documents++
		var fragments []model.Fragment
		if err := s.store.List(ctx, "fragment", document.ID, &fragments); err != nil {
			return report, err
		}
		coverage.Fragments += len(fragments)
		for _, fragment := range fragments {
			hits := matcher.Scan([]model.Fragment{fragment}, terms)
			coverage.Hits += len(hits)
			for _, hit := range hits {
				report.Hits++
				if hit.Forbidden || hit.Surface != preferredForHit(hit, terms) {
					coverage.Problematic++
					report.ProblematicHits++
				}
			}
		}
	}
	var suggestions []model.Suggestion
	if err := s.store.All(ctx, "suggestion", &suggestions); err != nil {
		return report, err
	}
	for _, suggestion := range suggestions {
		if suggestion.Status == model.SuggestionOpen && belongsToLibrary(suggestion, documents) {
			report.Suggestions++
		}
	}
	for _, coverage := range byLanguage {
		if coverage.Hits > 0 {
			coverage.CoverageRatio = float64(coverage.Hits-coverage.Problematic) / float64(coverage.Hits)
		}
		report.Languages = append(report.Languages, *coverage)
	}
	sort.Slice(report.Languages, func(i, j int) bool { return report.Languages[i].Language < report.Languages[j].Language })
	return report, nil
}

func preferredForHit(hit model.Hit, terms []model.Term) string {
	for _, term := range terms {
		if term.ID == hit.TermID {
			return term.Preferred
		}
	}
	return ""
}

func belongsToLibrary(suggestion model.Suggestion, documents []model.Document) bool {
	for _, document := range documents {
		if document.ID == suggestion.DocumentID {
			return true
		}
	}
	return false
}

func (s *Service) ExportSnapshot(ctx context.Context, libraryID string) (model.ExportSnapshot, error) {
	library, terms, err := s.LibrarySnapshot(ctx, libraryID)
	if err != nil {
		return model.ExportSnapshot{}, err
	}
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return model.ExportSnapshot{}, err
	}
	var tasks []model.CheckTask
	if err := s.store.List(ctx, "task", libraryID, &tasks); err != nil {
		return model.ExportSnapshot{}, err
	}
	suggestions := make([]model.Suggestion, 0)
	for _, task := range tasks {
		items, err := s.Suggestions(ctx, task.ID)
		if err != nil {
			return model.ExportSnapshot{}, err
		}
		suggestions = append(suggestions, items...)
	}
	audit, err := s.AuditTrail(ctx, libraryID, "")
	if err != nil {
		return model.ExportSnapshot{}, err
	}
	return model.ExportSnapshot{Library: library, Terms: terms, Documents: documents, Tasks: tasks, Suggestions: suggestions, Audit: audit, ExportedAt: s.now().UTC()}, nil
}
