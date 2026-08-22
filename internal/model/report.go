package model

import "time"

type LibraryReport struct {
	Library         Library   `json:"library"`
	Documents       int       `json:"documents"`
	Checked         int       `json:"checked"`
	Tasks           int       `json:"tasks"`
	OpenSuggestions int       `json:"open_suggestions"`
	Accepted        int       `json:"accepted"`
	Ignored         int       `json:"ignored"`
	GeneratedAt     time.Time `json:"generated_at"`
}
type TaskReport struct {
	Task        CheckTask `json:"task"`
	Suggestions int       `json:"suggestions"`
	Open        int       `json:"open"`
	Accepted    int       `json:"accepted"`
	Ignored     int       `json:"ignored"`
	Expired     int       `json:"expired"`
	Concepts    []string  `json:"concepts"`
}
type AuditEntry struct {
	SuggestionID string           `json:"suggestion_id"`
	Status       SuggestionStatus `json:"status"`
	At           time.Time        `json:"at"`
	Message      string           `json:"message"`
}
