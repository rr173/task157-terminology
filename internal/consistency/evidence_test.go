package consistency

import (
	"testing"

	"terminology/internal/model"
)

func TestRankOpenSuggestionsFirst(t *testing.T) {
	items := Rank([]model.Suggestion{{Concept: "a", Status: model.SuggestionAccepted}, {Concept: "z", Status: model.SuggestionOpen}})
	if items[0].Status != model.SuggestionOpen {
		t.Fatalf("open item should be actionable first: %#v", items)
	}
}

func TestEvidenceExplainsForbiddenSurface(t *testing.T) {
	evidence := BuildEvidence(model.Suggestion{Concept: "save", Actual: "存储", Expected: "保存", DocumentID: "d", FragmentID: "f"}, model.Fragment{ID: "f", Text: "请存储"}, model.Term{Concept: "save", Preferred: "保存", Forbidden: []string{"存储"}})
	if len(evidence.Reasons) < 2 {
		t.Fatalf("expected explicit reason: %#v", evidence)
	}
}
