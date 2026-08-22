package service

import (
	"context"
	"sort"
	"terminology/internal/model"
)

func (s *Service) LibraryReport(ctx context.Context, id string) (model.LibraryReport, error) {
	lib, e := s.Library(ctx, id)
	if e != nil {
		return model.LibraryReport{}, e
	}
	docs, e := s.Documents(ctx, id)
	if e != nil {
		return model.LibraryReport{}, e
	}
	var tasks []model.CheckTask
	if e = s.store.List(ctx, "task", id, &tasks); e != nil {
		return model.LibraryReport{}, e
	}
	r := model.LibraryReport{Library: lib, Documents: len(docs), Tasks: len(tasks), GeneratedAt: s.now().UTC()}
	for _, d := range docs {
		if d.Status == model.DocumentChecked {
			r.Checked++
		}
	}
	for _, t := range tasks {
		x, e := s.Suggestions(ctx, t.ID)
		if e != nil {
			return r, e
		}
		for _, v := range x {
			if v.Status.Reviewable() {
				r.OpenSuggestions++
				continue
			}
			switch v.Status {
			case model.SuggestionAccepted:
				r.Accepted++
			case model.SuggestionIgnored:
				r.Ignored++
			}
		}
	}
	return r, nil
}
func (s *Service) TaskReport(ctx context.Context, id string) (model.TaskReport, error) {
	var task model.CheckTask
	if e := s.store.Get(ctx, "task", id, &task); e != nil {
		return model.TaskReport{}, e
	}
	items, e := s.Suggestions(ctx, id)
	if e != nil {
		return model.TaskReport{}, e
	}
	r := model.TaskReport{Task: task, Suggestions: len(items)}
	seen := map[string]bool{}
	for _, v := range items {
		if !seen[v.Concept] {
			seen[v.Concept] = true
			r.Concepts = append(r.Concepts, v.Concept)
		}
		switch v.Status {
		case model.SuggestionOpen:
			r.Open++
		case model.SuggestionAccepted:
			r.Accepted++
		case model.SuggestionIgnored:
			r.Ignored++
		case model.SuggestionExpired:
			r.Expired++
		}
	}
	sort.Strings(r.Concepts)
	return r, nil
}
