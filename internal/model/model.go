package model

import "time"

type LibraryStatus string

const (
	LibraryDraft     LibraryStatus = "draft"
	LibraryPublished LibraryStatus = "published"
	LibraryRetired   LibraryStatus = "retired"
)

type DocumentStatus string

const (
	DocumentImported   DocumentStatus = "imported"
	DocumentChecked    DocumentStatus = "checked"
	DocumentSuperseded DocumentStatus = "superseded"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
)

type SuggestionStatus string

const (
	SuggestionOpen     SuggestionStatus = "open"
	SuggestionAccepted SuggestionStatus = "accepted"
	SuggestionIgnored  SuggestionStatus = "ignored"
	SuggestionExpired  SuggestionStatus = "expired"
)

type Library struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Version   int           `json:"version"`
	Status    LibraryStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}
type Term struct {
	ID        string    `json:"id"`
	LibraryID string    `json:"library_id"`
	Concept   string    `json:"concept"`
	Language  string    `json:"language"`
	Preferred string    `json:"preferred"`
	Forbidden []string  `json:"forbidden"`
	CreatedAt time.Time `json:"created_at"`
}
type Document struct {
	ID          string         `json:"id"`
	LibraryID   string         `json:"library_id"`
	Fingerprint string         `json:"fingerprint"`
	ExternalID  string         `json:"external_id"`
	Language    string         `json:"language"`
	Version     int            `json:"version"`
	Status      DocumentStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}
type Fragment struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Key        string    `json:"key"`
	Text       string    `json:"text"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
}
type Hit struct {
	ID         string    `json:"id"`
	FragmentID string    `json:"fragment_id"`
	TermID     string    `json:"term_id"`
	Concept    string    `json:"concept"`
	Language   string    `json:"language"`
	Surface    string    `json:"surface"`
	Start      int       `json:"start"`
	End        int       `json:"end"`
	Forbidden  bool      `json:"forbidden"`
	CreatedAt  time.Time `json:"created_at"`
}
type CheckTask struct {
	ID          string     `json:"id"`
	LibraryID   string     `json:"library_id"`
	DocumentIDs []string   `json:"document_ids"`
	Status      TaskStatus `json:"status"`
	Snapshot    string     `json:"snapshot"`
	CreatedAt   time.Time  `json:"created_at"`
	FinishedAt  time.Time  `json:"finished_at"`
}
type Suggestion struct {
	ID         string           `json:"id"`
	TaskID     string           `json:"task_id"`
	DocumentID string           `json:"document_id"`
	FragmentID string           `json:"fragment_id"`
	Concept    string           `json:"concept"`
	Language   string           `json:"language"`
	Actual     string           `json:"actual"`
	Expected   string           `json:"expected"`
	Status     SuggestionStatus `json:"status"`
	Evidence   []string         `json:"evidence"`
	CreatedAt  time.Time        `json:"created_at"`
	ReviewedAt time.Time        `json:"reviewed_at"`
}
type Diff struct {
	EarlierID              string `json:"earlier_id"`
	LaterID                string `json:"later_id"`
	AddedFragments         int    `json:"added_fragments"`
	RemovedFragments       int    `json:"removed_fragments"`
	ChangedFragments       int    `json:"changed_fragments"`
	InvalidatedSuggestions int    `json:"invalidated_suggestions"`
}
