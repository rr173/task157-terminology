package model

import "fmt"

func (s SuggestionStatus) Reviewable() bool { return s == SuggestionOpen }
func (s TaskStatus) Terminal() bool         { return s == TaskSucceeded || s == TaskFailed }
// Current reports whether a document is still the live version for its external
// id. A superseded document has been replaced by a newer version and must not be
// re-checked, reported, or counted as part of the current corpus.
func (s DocumentStatus) Current() bool {
	return s == DocumentImported || s == DocumentChecked
}
func TransitionTask(a, b TaskStatus) error {
	ok := map[TaskStatus]map[TaskStatus]bool{TaskPending: {TaskRunning: true}, TaskRunning: {TaskSucceeded: true, TaskFailed: true}, TaskFailed: {TaskPending: true}}
	if a == b || ok[a][b] {
		return nil
	}
	return fmt.Errorf("task transition %s -> %s is not allowed", a, b)
}
