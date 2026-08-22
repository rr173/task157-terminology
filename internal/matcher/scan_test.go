package matcher

import (
	"testing"

	"github.com/rr173/task157-terminology/internal/model"
)

func TestHitsFindEveryForbiddenOccurrence(t *testing.T) {
	fragment := model.Fragment{ID: "f", Text: "存储后再次存储"}
	hits := Hits(fragment, []model.Term{{ID: "t", Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}}})
	if len(hits) != 2 || hits[0].Start == hits[1].Start {
		t.Fatalf("expected two coordinates, got %#v", hits)
	}
}

func TestOccurrencesIgnoresEmptyPhrase(t *testing.T) {
	if got := Occurrences("abc", ""); got != nil {
		t.Fatalf("empty phrase should not match: %#v", got)
	}
}
