package service

import (
	"context"
	"terminology/internal/model"
)

func (s *Service) Diff(ctx context.Context, earlier, later string) (model.Diff, error) {
	var a, b model.Document
	if e := s.store.Get(ctx, "document", earlier, &a); e != nil {
		return model.Diff{}, e
	}
	if e := s.store.Get(ctx, "document", later, &b); e != nil {
		return model.Diff{}, e
	}
	var af, bf []model.Fragment
	if e := s.store.List(ctx, "fragment", a.ID, &af); e != nil {
		return model.Diff{}, e
	}
	if e := s.store.List(ctx, "fragment", b.ID, &bf); e != nil {
		return model.Diff{}, e
	}
	left := map[string]model.Fragment{}
	right := map[string]model.Fragment{}
	for _, v := range af {
		left[v.Key] = v
	}
	for _, v := range bf {
		right[v.Key] = v
	}
	out := model.Diff{EarlierID: a.ID, LaterID: b.ID}
	for k, v := range right {
		if old, ok := left[k]; !ok {
			out.AddedFragments++
		} else if old.Text != v.Text {
			out.ChangedFragments++
		}
	}
	for k := range left {
		if _, ok := right[k]; !ok {
			out.RemovedFragments++
		}
	}
	count, err := s.expiredSuggestionsForDocument(ctx, b.ID)
	if err != nil {
		return out, err
	}
	out.InvalidatedSuggestions = count
	return out, nil
}
