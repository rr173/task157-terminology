package model

import "testing"

func TestTaskTransition(t *testing.T) {
	if e := TransitionTask(TaskPending, TaskRunning); e != nil {
		t.Fatal(e)
	}
	if e := TransitionTask(TaskSucceeded, TaskRunning); e == nil {
		t.Fatal("terminal task restarted")
	}
}
