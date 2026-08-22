package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rr173/task157-terminology/internal/httpapi"
	"github.com/rr173/task157-terminology/internal/model"
	"github.com/rr173/task157-terminology/internal/service"
	"github.com/rr173/task157-terminology/internal/store"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	db := flag.String("db", "terminology.db", "")
	addr := flag.String("addr", ":8080", "")
	smoke := flag.Bool("smoke-test", false, "")
	flag.Parse()
	if *smoke {
		runSmoke()
		return
	}
	s, e := store.Open(*db)
	if e != nil {
		log.Fatal(e)
	}
	defer s.Close()
	core := service.New(s)
	_, _ = core.Recover(context.Background())
	m := http.NewServeMux()
	m.Handle("/v1/", httpapi.New(core).Handler())
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		report, err := core.Health(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	})
	log.Fatal(http.ListenAndServe(*addr, m))
}
func runSmoke() {
	dir, e := os.MkdirTemp("", "term")
	if e != nil {
		log.Fatal(e)
	}
	defer os.RemoveAll(dir)
	db, e := store.Open(filepath.Join(dir, "s.db"))
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	s := service.New(db)
	ctx := context.Background()
	lib, e := s.CreateLibrary(ctx, model.LibraryInput{Name: "demo"})
	if e != nil {
		log.Fatal(e)
	}
	_, e = s.AddTerm(ctx, lib.ID, model.TermInput{Concept: "save", Language: "zh", Preferred: "保存", Forbidden: []string{"存储"}})
	if e != nil {
		log.Fatal(e)
	}
	_, e = s.Publish(ctx, lib.ID)
	if e != nil {
		log.Fatal(e)
	}
	doc, _, e := s.ImportDocument(ctx, lib.ID, model.DocumentInput{ExternalID: "d", Language: "zh", Fingerprint: "a", Fragments: []model.FragmentInput{{Key: "p", Text: "请存储文件", Position: 1}}})
	if e != nil {
		log.Fatal(e)
	}
	_, _, e = s.Check(ctx, lib.ID, []string{doc.ID})
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println("smoke-test passed")
}
