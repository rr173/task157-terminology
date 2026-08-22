package consistency

import (
	"terminology/internal/model"
	"testing"
)

func TestGroup(t *testing.T) {
	if len(Group([]model.Suggestion{{Concept: "save", Actual: "存储"}})) != 1 {
		t.Fatal("group failed")
	}
}
