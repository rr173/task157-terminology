package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"terminology/internal/model"
)

func (s *Service) recordAudit(ctx context.Context, libraryID, entityType, entityID, action, actor, message string, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	record := model.AuditRecord{ID: uuid.NewString(), LibraryID: libraryID, EntityType: entityType, EntityID: entityID, Action: action, Actor: actor, Message: message, Metadata: metadata, CreatedAt: s.now().UTC()}
	return s.store.Save(ctx, "audit", record.ID, libraryID, int(record.CreatedAt.UnixNano()), record)
}

func (s *Service) AuditTrail(ctx context.Context, libraryID string, entityType string) ([]model.AuditRecord, error) {
	var records []model.AuditRecord
	if err := s.store.List(ctx, "audit", libraryID, &records); err != nil {
		return nil, err
	}
	filtered := records[:0]
	for _, record := range records {
		if entityType == "" || record.EntityType == entityType {
			filtered = append(filtered, record)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.Before(filtered[j].CreatedAt) })
	return filtered, nil
}

func (s *Service) ReviewWithActor(ctx context.Context, id string, accept bool, actor, reason string) (model.Suggestion, model.ReviewDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var suggestion model.Suggestion
	if err := s.store.Require(ctx, "suggestion", id, &suggestion); err != nil {
		return suggestion, model.ReviewDecision{}, err
	}
	if suggestion.Status == model.SuggestionExpired {
		return suggestion, model.ReviewDecision{}, fmt.Errorf("suggestion is expired")
	}
	if suggestion.Status != model.SuggestionOpen {
		return suggestion, model.ReviewDecision{}, fmt.Errorf("suggestion is already reviewed")
	}
	before := suggestion.Status
	if accept {
		suggestion.Status = model.SuggestionAccepted
	} else {
		suggestion.Status = model.SuggestionIgnored
	}
	suggestion.ReviewedAt = s.now().UTC()
	decision := model.ReviewDecision{SuggestionID: id, Before: before, After: suggestion.Status, Actor: actor, Reason: reason, At: suggestion.ReviewedAt}
	if err := decision.Validate(); err != nil {
		return suggestion, decision, err
	}
	if err := s.store.Save(ctx, "suggestion", suggestion.ID, suggestion.TaskID, 0, suggestion); err != nil {
		return suggestion, decision, err
	}
	var task model.CheckTask
	if err := s.store.Get(ctx, "task", suggestion.TaskID, &task); err == nil {
		_ = s.recordAudit(ctx, task.LibraryID, "suggestion", suggestion.ID, "review", actor, reason, map[string]any{"before": before, "after": suggestion.Status})
	}
	return suggestion, decision, nil
}
