package service

import (
	"context"
	"terminology/internal/model"
)

func (s *Service) Suggestions(ctx context.Context, taskID string) ([]model.Suggestion, error) {
	var v []model.Suggestion
	return v, s.store.List(ctx, "suggestion", taskID, &v)
}
func (s *Service) Review(ctx context.Context, id string, accept bool) (model.Suggestion, error) {
	v, _, err := s.ReviewWithActor(ctx, id, accept, "api", "reviewed through the default endpoint")
	return v, err
}
func (s *Service) Recover(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var tasks []model.CheckTask
	if err := s.store.All(ctx, "task", &tasks); err != nil {
		return nil, err
	}
	out := []string{}
	for _, task := range tasks {
		if task.Status == model.TaskRunning {
			if err := model.TransitionTask(task.Status, model.TaskPending); err != nil {
				return out, err
			}
			task.Status = model.TaskPending
			if err := s.store.Save(ctx, "task", task.ID, task.LibraryID, 0, task); err != nil {
				return out, err
			}
			out = append(out, task.ID)
		}
	}
	return out, nil
}
