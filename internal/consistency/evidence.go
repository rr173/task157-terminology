package consistency

import (
	"github.com/rr173/task157-terminology/internal/matcher"
	"github.com/rr173/task157-terminology/internal/model"
	"sort"
)

type Evidence struct {
	Concept    string   `json:"concept"`
	DocumentID string   `json:"document_id"`
	FragmentID string   `json:"fragment_id"`
	Language   string   `json:"language"`
	Actual     string   `json:"actual"`
	Expected   string   `json:"expected"`
	Context    string   `json:"context"`
	Reasons    []string `json:"reasons"`
}

func BuildEvidence(suggestion model.Suggestion, fragment model.Fragment, term model.Term) Evidence {
	reasons := []string{"translation differs from the published preferred form"}
	if suggestion.Actual != term.Preferred {
		reasons = append(reasons, "surface is not preferred")
	}
	for _, forbidden := range term.Forbidden {
		if suggestion.Actual == forbidden {
			reasons = append(reasons, "surface is explicitly forbidden")
			break
		}
	}
	return Evidence{Concept: suggestion.Concept, DocumentID: suggestion.DocumentID, FragmentID: suggestion.FragmentID, Language: suggestion.Language, Actual: suggestion.Actual, Expected: suggestion.Expected, Context: matcher.Context(fragment.Text, 0, len(fragment.Text)), Reasons: reasons}
}

func Rank(items []model.Suggestion) []model.Suggestion {
	result := append([]model.Suggestion(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Status != result[j].Status {
			return result[i].Status == model.SuggestionOpen
		}
		if result[i].Concept != result[j].Concept {
			return result[i].Concept < result[j].Concept
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

func Concepts(items []model.Suggestion) []string {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Concept] = true
	}
	result := make([]string, 0, len(seen))
	for concept := range seen {
		result = append(result, concept)
	}
	sort.Strings(result)
	return result
}
