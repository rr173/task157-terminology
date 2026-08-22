package service

import (
	"context"

	"terminology/internal/model"
)

func (s *Service) expireOpenSuggestionsForDocument(ctx context.Context, documentID string) (int, error) {
	var tasks []model.CheckTask
	if err := s.store.All(ctx, "task", &tasks); err != nil {
		return 0, err
	}
	expired := 0
	for _, task := range tasks {
		items, err := s.Suggestions(ctx, task.ID)
		if err != nil {
			return expired, err
		}
		for _, item := range items {
			if item.DocumentID != documentID || item.Status != model.SuggestionOpen {
				continue
			}
			item.Status = model.SuggestionExpired
			item.ReviewedAt = s.now().UTC()
			if err := s.store.Save(ctx, "suggestion", item.ID, item.TaskID, 0, item); err != nil {
				return expired, err
			}
			expired++
		}
	}
	return expired, nil
}

func (s *Service) expiredSuggestionsForDocument(ctx context.Context, documentID string) (int, error) {
	var tasks []model.CheckTask
	if err := s.store.All(ctx, "task", &tasks); err != nil {
		return 0, err
	}
	count := 0
	for _, task := range tasks {
		items, err := s.Suggestions(ctx, task.ID)
		if err != nil {
			return count, err
		}
		for _, item := range items {
			if item.DocumentID == documentID && item.Status == model.SuggestionExpired {
				count++
			}
		}
	}
	return count, nil
}
