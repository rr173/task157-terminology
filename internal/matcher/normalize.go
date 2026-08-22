package matcher

import (
	"strings"
	"unicode"
)

func Normalize(text string) string {
	var out strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(text)) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			out.WriteRune(' ')
			continue
		}
		out.WriteRune(r)
	}
	return strings.Join(strings.Fields(out.String()), " ")
}
func Contains(text, phrase string) bool { return strings.Contains(Normalize(text), Normalize(phrase)) }
func Context(text string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	left := start - 24
	if left < 0 {
		left = 0
	}
	right := end + 24
	if right > len(text) {
		right = len(text)
	}
	return strings.TrimSpace(text[left:right])
}
