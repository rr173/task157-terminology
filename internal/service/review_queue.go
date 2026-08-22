package service

import (
	"context"
	"sort"
	"strings"

	"terminology/internal/model"
)

func (s *Service) ReviewQueue(ctx context.Context, query model.ReviewQueueQuery) (model.ReviewQueue, error) {
	if _, err := s.Library(ctx, query.LibraryID); err != nil {
		return model.ReviewQueue{}, err
	}
	tasks, err := s.libraryTasks(ctx, query.LibraryID)
	if err != nil {
		return model.ReviewQueue{}, err
	}
	queue := model.ReviewQueue{
		LibraryID: query.LibraryID,
		Items:     []model.Suggestion{},
		ByStatus: map[model.SuggestionStatus]int{
			model.SuggestionOpen: 0, model.SuggestionAccepted: 0,
			model.SuggestionIgnored: 0, model.SuggestionExpired: 0,
		},
		GeneratedAt: s.now().UTC(),
	}
	for _, task := range tasks {
		items, listErr := s.Suggestions(ctx, task.ID)
		if listErr != nil {
			return queue, listErr
		}
		for _, item := range items {
			if !matchesReviewQueue(item, query) {
				continue
			}
			queue.Total++
			queue.ByStatus[item.Status]++
			queue.Items = append(queue.Items, item)
		}
	}
	sort.Slice(queue.Items, func(i, j int) bool {
		if queue.Items[i].Status != queue.Items[j].Status {
			return reviewPriority(queue.Items[i].Status) < reviewPriority(queue.Items[j].Status)
		}
		if !queue.Items[i].CreatedAt.Equal(queue.Items[j].CreatedAt) {
			return queue.Items[i].CreatedAt.Before(queue.Items[j].CreatedAt)
		}
		return queue.Items[i].ID < queue.Items[j].ID
	})
	queue.Items, queue.NextOffset = reviewQueuePage(queue.Items, query.Offset, query.PageSize())
	return queue, nil
}

func (s *Service) libraryTasks(ctx context.Context, libraryID string) ([]model.CheckTask, error) {
	var tasks []model.CheckTask
	if err := s.store.List(ctx, "task", libraryID, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func matchesReviewQueue(item model.Suggestion, query model.ReviewQueueQuery) bool {
	if query.Status != "" {
		if query.Status == model.SuggestionOpen {
			if !item.Status.Reviewable() {
				return false
			}
		} else if item.Status != query.Status {
			return false
		}
	}
	if language := strings.TrimSpace(query.Language); language != "" && item.Language != language {
		return false
	}
	if concept := strings.TrimSpace(query.Concept); concept != "" && !strings.EqualFold(item.Concept, concept) {
		return false
	}
	return true
}

func reviewPriority(status model.SuggestionStatus) int {
	switch status {
	case model.SuggestionOpen:
		return 0
	case model.SuggestionExpired:
		return 1
	case model.SuggestionAccepted:
		return 2
	case model.SuggestionIgnored:
		return 3
	default:
		return 4
	}
}

func reviewQueuePage(items []model.Suggestion, offset, limit int) ([]model.Suggestion, int) {
	if offset >= len(items) {
		return []model.Suggestion{}, 0
	}
	end := offset + limit
	next := 0
	if end < len(items) {
		next = end
	} else {
		end = len(items)
	}
	return append([]model.Suggestion(nil), items[offset:end]...), next
}
