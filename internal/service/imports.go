package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/rr173/task157-terminology/internal/matcher"
	"github.com/rr173/task157-terminology/internal/model"
)

func (s *Service) ImportBatch(ctx context.Context, libraryID string, input model.ImportBatchInput) (model.ImportBatch, error) {
	if err := input.Validate(); err != nil {
		return model.ImportBatch{}, err
	}
	library, err := s.Library(ctx, libraryID)
	if err != nil {
		return model.ImportBatch{}, err
	}
	if library.Status != model.LibraryPublished {
		return model.ImportBatch{}, fmt.Errorf("library must be published")
	}
	batch := model.ImportBatch{ID: uuid.NewString(), LibraryID: libraryID, Source: strings.TrimSpace(input.Source), Status: model.ImportBatchRunning, CreatedAt: s.now().UTC()}
	if err := s.store.Save(ctx, "batch", batch.ID, libraryID, 0, batch); err != nil {
		return batch, err
	}
	for _, documentInput := range input.Documents {
		document, replayed, importErr := s.ImportDocument(ctx, libraryID, documentInput)
		if importErr != nil {
			batch.Rejected++
			batch.Errors = append(batch.Errors, fmt.Sprintf("%s: %v", documentInput.ExternalID, importErr))
			continue
		}
		batch.DocumentIDs = append(batch.DocumentIDs, document.ID)
		if replayed {
			batch.Replayed++
		} else {
			batch.Accepted++
		}
	}
	batch.FinishedAt = s.now().UTC()
	switch {
	case batch.Rejected == len(input.Documents):
		batch.Status = model.ImportBatchFailed
	case batch.Rejected > 0:
		batch.Status = model.ImportBatchPartial
	default:
		batch.Status = model.ImportBatchDone
	}
	if err := s.store.Save(ctx, "batch", batch.ID, libraryID, 0, batch); err != nil {
		return batch, err
	}
	_ = s.recordAudit(ctx, libraryID, "batch", batch.ID, "import", input.Source, "document batch processed", map[string]any{"accepted": batch.Accepted, "replayed": batch.Replayed, "rejected": batch.Rejected})
	return batch, nil
}

func (s *Service) Batches(ctx context.Context, libraryID string) ([]model.ImportBatch, error) {
	var batches []model.ImportBatch
	if err := s.store.List(ctx, "batch", libraryID, &batches); err != nil {
		return nil, err
	}
	sort.Slice(batches, func(i, j int) bool { return batches[i].CreatedAt.After(batches[j].CreatedAt) })
	return batches, nil
}

func (s *Service) CurrentDocuments(ctx context.Context, libraryID, language string) ([]model.Document, error) {
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	result := documents[:0]
	for _, document := range documents {
		if document.Status == model.DocumentSuperseded {
			continue
		}
		if language != "" && !strings.EqualFold(language, document.Language) {
			continue
		}
		result = append(result, document)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ExternalID == result[j].ExternalID {
			return result[i].Version > result[j].Version
		}
		return result[i].ExternalID < result[j].ExternalID
	})
	return result, nil
}

func (s *Service) DocumentHistory(ctx context.Context, libraryID, externalID string) ([]model.Document, error) {
	documents, err := s.Documents(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	result := documents[:0]
	for _, document := range documents {
		if document.ExternalID == externalID {
			result = append(result, document)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Version < result[j].Version })
	return result, nil
}

func (s *Service) SearchDocuments(ctx context.Context, query model.DocumentQuery) ([]model.DocumentSearchResult, error) {
	documents, err := s.CurrentDocuments(ctx, query.LibraryID, query.Language)
	if err != nil {
		return nil, err
	}
	needle := matcher.Normalize(query.Text)
	result := make([]model.DocumentSearchResult, 0)
	for _, document := range documents {
		if query.External != "" && document.ExternalID != query.External {
			continue
		}
		if query.Status != "" && string(document.Status) != query.Status {
			continue
		}
		var fragments []model.Fragment
		if err := s.store.List(ctx, "fragment", document.ID, &fragments); err != nil {
			return nil, err
		}
		matches := make([]string, 0)
		for _, fragment := range fragments {
			if needle == "" || strings.Contains(matcher.Normalize(fragment.Text), needle) {
				matches = append(matches, fragment.Key)
			}
		}
		if needle == "" || len(matches) > 0 {
			result = append(result, model.DocumentSearchResult{Document: document, Matches: matches})
		}
		if len(result) >= query.NormalizedLimit() {
			break
		}
	}
	return result, nil
}
