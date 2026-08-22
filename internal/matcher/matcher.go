package matcher

import (
	"github.com/rr173/task157-terminology/internal/model"
)

func Hits(fragment model.Fragment, terms []model.Term) []model.Hit {
	out := []model.Hit{}
	for _, term := range terms {
		surfaces := append([]string{term.Preferred}, term.Forbidden...)
		for _, surface := range surfaces {
			for _, span := range Occurrences(fragment.Text, surface) {
				out = append(out, model.Hit{FragmentID: fragment.ID, TermID: term.ID, Concept: term.Concept, Language: term.Language, Surface: surface, Start: span.Start, End: span.End, Forbidden: surface != term.Preferred, CreatedAt: model.Now()})
			}
		}
	}
	return out
}
