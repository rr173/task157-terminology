package model

import (
	"fmt"
	"strings"
	"time"
)

// ImportBatch records one client submission.  Keeping the batch separate from
// documents lets callers retry a network request without creating a second
// logical import and gives operators a durable unit to inspect after restart.
type ImportBatch struct {
	ID          string            `json:"id"`
	LibraryID   string            `json:"library_id"`
	Source      string            `json:"source"`
	Status      ImportBatchStatus `json:"status"`
	DocumentIDs []string          `json:"document_ids"`
	Accepted    int               `json:"accepted"`
	Replayed    int               `json:"replayed"`
	Rejected    int               `json:"rejected"`
	Errors      []string          `json:"errors"`
	CreatedAt   time.Time         `json:"created_at"`
	FinishedAt  time.Time         `json:"finished_at"`
}

type ImportBatchStatus string

const (
	ImportBatchRunning ImportBatchStatus = "running"
	ImportBatchDone    ImportBatchStatus = "done"
	ImportBatchPartial ImportBatchStatus = "partial"
	ImportBatchFailed  ImportBatchStatus = "failed"
)

func (b ImportBatch) Terminal() bool {
	return b.Status == ImportBatchDone || b.Status == ImportBatchPartial || b.Status == ImportBatchFailed
}

func (b ImportBatch) Successful() bool {
	return b.Status == ImportBatchDone || b.Status == ImportBatchPartial
}

type ReviewDecision struct {
	SuggestionID string           `json:"suggestion_id"`
	Before       SuggestionStatus `json:"before"`
	After        SuggestionStatus `json:"after"`
	Actor        string           `json:"actor"`
	Reason       string           `json:"reason"`
	At           time.Time        `json:"at"`
}

type AuditRecord struct {
	ID         string         `json:"id"`
	LibraryID  string         `json:"library_id"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Action     string         `json:"action"`
	Actor      string         `json:"actor"`
	Message    string         `json:"message"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (v ReviewDecision) Validate() error {
	if strings.TrimSpace(v.SuggestionID) == "" {
		return fmt.Errorf("suggestion id is required")
	}
	if v.Before == v.After {
		return fmt.Errorf("review must change suggestion status")
	}
	if v.After != SuggestionAccepted && v.After != SuggestionIgnored {
		return fmt.Errorf("review result must be accepted or ignored")
	}
	return nil
}

type DocumentQuery struct {
	LibraryID string `json:"library_id"`
	Language  string `json:"language"`
	External  string `json:"external_id"`
	Status    string `json:"status"`
	Text      string `json:"text"`
	Limit     int    `json:"limit"`
}

func (q DocumentQuery) NormalizedLimit() int {
	if q.Limit <= 0 || q.Limit > 200 {
		return 50
	}
	return q.Limit
}

type DocumentSearchResult struct {
	Document Document `json:"document"`
	Matches  []string `json:"matches"`
}

type ImportBatchInput struct {
	Source    string          `json:"source"`
	Documents []DocumentInput `json:"documents"`
}

func (v ImportBatchInput) Validate() error {
	if strings.TrimSpace(v.Source) == "" {
		return fmt.Errorf("batch source is required")
	}
	if len(v.Documents) == 0 {
		return fmt.Errorf("batch must contain documents")
	}
	// Individual documents are validated inside ImportBatch so that a single
	// malformed entry is reported as a per-document rejection rather than
	// aborting the whole submission.
	return nil
}
