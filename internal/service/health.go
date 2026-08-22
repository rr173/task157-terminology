package service

import (
	"context"
	"terminology/internal/model"
)

func (s *Service) Health(ctx context.Context) (map[string]any, error) {
	report, err := s.HealthReport(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{"libraries": report.Libraries, "documents": report.Documents, "running_tasks": report.RunningTasks, "open_reviews": report.OpenReviews, "healthy": report.Healthy, "checked_at": report.CheckedAt}, nil
}

func (s *Service) HealthReport(ctx context.Context) (model.HealthReport, error) {
	var libraries []model.Library
	if err := s.store.All(ctx, "library", &libraries); err != nil {
		return model.HealthReport{}, err
	}
	var documents []model.Document
	if err := s.store.All(ctx, "document", &documents); err != nil {
		return model.HealthReport{}, err
	}
	var tasks []model.CheckTask
	if err := s.store.All(ctx, "task", &tasks); err != nil {
		return model.HealthReport{}, err
	}
	var suggestions []model.Suggestion
	if err := s.store.All(ctx, "suggestion", &suggestions); err != nil {
		return model.HealthReport{}, err
	}
	report := model.HealthReport{Healthy: true, Libraries: len(libraries), Documents: len(documents), CheckedAt: s.now().UTC()}
	for _, task := range tasks {
		if task.Status == model.TaskRunning {
			report.RunningTasks++
		}
	}
	for _, suggestion := range suggestions {
		if suggestion.Status.Reviewable() {
			report.OpenReviews++
		}
	}
	return report, nil
}
func (s *Service) DocumentsForTask(ctx context.Context, id string) ([]model.Document, error) {
	var t model.CheckTask
	if e := s.store.Get(ctx, "task", id, &t); e != nil {
		return nil, e
	}
	out := []model.Document{}
	for _, docID := range t.DocumentIDs {
		var d model.Document
		if e := s.store.Get(ctx, "document", docID, &d); e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, nil
}
