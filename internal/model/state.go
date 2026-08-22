package model

import "fmt"

func (s SuggestionStatus) Reviewable() bool { return s == SuggestionOpen }
func (s TaskStatus) Terminal() bool         { return s == TaskSucceeded || s == TaskFailed }
func (s DocumentStatus) Current() bool      { return s != "" }
func TransitionTask(a, b TaskStatus) error {
	ok := map[TaskStatus]map[TaskStatus]bool{TaskPending: {TaskRunning: true}, TaskRunning: {TaskSucceeded: true, TaskFailed: true}, TaskFailed: {TaskPending: true}}
	if a == b || ok[a][b] {
		return nil
	}
	return fmt.Errorf("task transition %s -> %s is not allowed", a, b)
}
