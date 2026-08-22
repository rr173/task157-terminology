package model

import (
	"fmt"
	"strings"
	"time"
)

// ReviewQueueQuery narrows the reviewer-facing queue without changing the
// immutable suggestion records produced by a completed check task.
type ReviewQueueQuery struct {
	LibraryID string           `json:"library_id"`
	Language  string           `json:"language"`
	Concept   string           `json:"concept"`
	Status    SuggestionStatus `json:"status"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

func (q ReviewQueueQuery) Validate() error {
	if strings.TrimSpace(q.LibraryID) == "" {
		return fmt.Errorf("library id is required")
	}
	if q.Offset < 0 {
		return fmt.Errorf("offset must not be negative")
	}
	return nil
}

func (q ReviewQueueQuery) PageSize() int {
	if q.Limit <= 0 {
		return 50
	}
	if q.Limit > 200 {
		return 200
	}
	return q.Limit
}

// ReviewQueue exposes the current workload for one terminology library. The
// status counters describe the full filtered result, while Items is one page.
type ReviewQueue struct {
	LibraryID   string                   `json:"library_id"`
	Items       []Suggestion             `json:"items"`
	Total       int                      `json:"total"`
	ByStatus    map[SuggestionStatus]int `json:"by_status"`
	NextOffset  int                      `json:"next_offset,omitempty"`
	GeneratedAt time.Time                `json:"generated_at"`
}

func (q ReviewQueue) HasNextPage() bool {
	return q.NextOffset > 0
}

func (q ReviewQueue) OpenCount() int {
	return q.ByStatus[SuggestionOpen]
}
