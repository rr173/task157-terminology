package matcher

import "testing"

func TestNormalize(t *testing.T) {
	if Normalize("  Save, File! ") != "save file" {
		t.Fatal(Normalize("  Save, File! "))
	}
}
