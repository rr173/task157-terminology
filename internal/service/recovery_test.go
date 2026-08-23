package service

import (
	"context"
	"path/filepath"
	"testing"

	"terminology/internal/model"
	"terminology/internal/store"
)

func TestRecoverRunningTaskAndPersistReviewAudit(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "recovery.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := New(db)
	ctx := context.Background()
	task := model.CheckTask{ID: "task-running", LibraryID: "library", Status: model.TaskRunning}
	if err := db.Save(ctx, "task", task.ID, task.LibraryID, 0, task); err != nil {
		t.Fatal(err)
	}
	ids, err := s.Recover(ctx)
	if err != nil || len(ids) != 1 {
		t.Fatalf("recovery failed: %#v %v", ids, err)
	}
	var restored model.CheckTask
	if err := db.Get(ctx, "task", task.ID, &restored); err != nil || restored.Status != model.TaskPending {
		t.Fatalf("task was not reset: %#v %v", restored, err)
	}
}

func TestImportBatchCountsReplay(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "batch.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := New(db)
	ctx := context.Background()
	lib, err := s.CreateLibrary(ctx, model.LibraryInput{Name: "batch"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.AddTerm(ctx, lib.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存"}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Publish(ctx, lib.ID); err != nil {
		t.Fatal(err)
	}
	input := model.ImportBatchInput{Source: "test", Documents: []model.DocumentInput{{ExternalID: "doc", Language: "zh", Fingerprint: "fp", Fragments: []model.FragmentInput{{Key: "p", Text: "保存", Position: 0}}}}}
	first, err := s.ImportBatch(ctx, lib.ID, input)
	if err != nil || first.Accepted != 1 {
		t.Fatalf("first import failed: %#v %v", first, err)
	}
	second, err := s.ImportBatch(ctx, lib.ID, input)
	if err != nil || second.Replayed != 1 {
		t.Fatalf("replay was not counted: %#v %v", second, err)
	}
}
