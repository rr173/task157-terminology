package matcher

import (
	"github.com/rr173/task157-terminology/internal/model"
	"sort"
)

func Scan(fragments []model.Fragment, terms []model.Term) []model.Hit {
	out := []model.Hit{}
	for _, f := range fragments {
		out = append(out, Hits(f, terms)...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FragmentID == out[j].FragmentID {
			return out[i].Start < out[j].Start
		}
		return out[i].FragmentID < out[j].FragmentID
	})
	return out
}
func Forbidden(hits []model.Hit) []model.Hit {
	out := []model.Hit{}
	for _, h := range hits {
		if h.Forbidden {
			out = append(out, h)
		}
	}
	return out
}
