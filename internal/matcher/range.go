package matcher

import (
	"strings"
	"unicode/utf8"
)

type Range struct {
	Start int
	End   int
}

// Occurrences returns every non-overlapping occurrence while preserving byte
// offsets, which makes the coordinates safe to apply to the original UTF-8
// fragment. Empty terms are deliberately ignored.
func Occurrences(text, phrase string) []Range {
	if phrase == "" || text == "" {
		return nil
	}
	needle := strings.ToLower(phrase)
	haystack := strings.ToLower(text)
	result := make([]Range, 0, 2)
	for offset := 0; offset < len(haystack); {
		idx := strings.Index(haystack[offset:], needle)
		if idx < 0 {
			break
		}
		start := offset + idx
		end := start + len(needle)
		if utf8.ValidString(text[start:end]) {
			result = append(result, Range{Start: start, End: end})
		}
		offset = end
	}
	return result
}

func MergeRanges(ranges []Range) []Range {
	if len(ranges) < 2 {
		return ranges
	}
	out := make([]Range, 0, len(ranges))
	for _, current := range ranges {
		if len(out) == 0 || current.Start > out[len(out)-1].End {
			out = append(out, current)
			continue
		}
		if current.End > out[len(out)-1].End {
			out[len(out)-1].End = current.End
		}
	}
	return out
}
