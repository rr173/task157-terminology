package consistency

import (
	"github.com/rr173/task157-terminology/internal/model"
	"sort"
)

type Summary struct {
	Concept   string
	Expected  string
	Actuals   []string
	Documents []string
}

func Group(items []model.Suggestion) []Summary {
	m := map[string]*Summary{}
	for _, v := range items {
		x := m[v.Concept]
		if x == nil {
			x = &Summary{Concept: v.Concept, Expected: v.Expected}
			m[v.Concept] = x
		}
		x.Actuals = append(x.Actuals, v.Actual)
		x.Documents = append(x.Documents, v.DocumentID)
	}
	out := []Summary{}
	for _, v := range m {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Concept < out[j].Concept })
	return out
}
