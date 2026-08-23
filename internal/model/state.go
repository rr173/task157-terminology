package model

import "fmt"

func (s SuggestionStatus) Reviewable() bool { return s == SuggestionOpen }

// Valid reports whether s is one of the supported suggestion status values.
// Callers must reject an invalid status rather than silently treating it as a
// filter that matches nothing, which would hide a misbehaving client.
func (s SuggestionStatus) Valid() bool {
	switch s {
	case SuggestionOpen, SuggestionAccepted, SuggestionIgnored, SuggestionExpired:
		return true
	}
	return false
}
func (s TaskStatus) Terminal() bool         { return s == TaskSucceeded || s == TaskFailed }
func (s DocumentStatus) Current() bool      { return s == DocumentImported || s == DocumentChecked }
func TransitionTask(a, b TaskStatus) error {
	ok := map[TaskStatus]map[TaskStatus]bool{TaskPending: {TaskRunning: true}, TaskRunning: {TaskSucceeded: true, TaskFailed: true}, TaskFailed: {TaskPending: true}}
	if a == b || ok[a][b] {
		return nil
	}
	return fmt.Errorf("task transition %s -> %s is not allowed", a, b)
}
