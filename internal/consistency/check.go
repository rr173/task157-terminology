package consistency

import (
	"terminology/internal/model"
)

func Suggestions(task model.CheckTask, docs []model.Document, fragments []model.Fragment, hits []model.Hit, terms []model.Term) []model.Suggestion {
	byTerm := map[string]model.Term{}
	for _, term := range terms {
		byTerm[term.ID] = term
	}
	docByFragment := map[string]model.Document{}
	for _, fragment := range fragments {
		for _, doc := range docs {
			if fragment.DocumentID == doc.ID {
				docByFragment[fragment.ID] = doc
			}
		}
	}
	out := []model.Suggestion{}
	for _, hit := range hits {
		term, ok := byTerm[hit.TermID]
		if !ok {
			continue
		}
		doc := docByFragment[hit.FragmentID]
		if hit.Forbidden || hit.Surface != term.Preferred {
			reasons := []string{"term match requires preferred translation"}
			if hit.Forbidden {
				reasons = append(reasons, "surface is forbidden")
			}
			out = append(out, model.Suggestion{TaskID: task.ID, DocumentID: doc.ID, FragmentID: hit.FragmentID, Concept: term.Concept, Language: doc.Language, Actual: hit.Surface, Expected: term.Preferred, Status: model.SuggestionOpen, Evidence: reasons, CreatedAt: model.Now()})
		}
	}
	return Rank(out)
}
