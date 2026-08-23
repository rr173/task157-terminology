package model

import (
	"fmt"
	"strings"
	"time"
)

type LibraryInput struct {
	Name string `json:"name"`
}
type TermInput struct {
	Concept   string   `json:"concept"`
	Language  string   `json:"language"`
	Preferred string   `json:"preferred"`
	Forbidden []string `json:"forbidden"`
}
type DocumentInput struct {
	ExternalID  string          `json:"external_id"`
	Language    string          `json:"language"`
	Fingerprint string          `json:"fingerprint"`
	Fragments   []FragmentInput `json:"fragments"`
}
type FragmentInput struct {
	Key      string `json:"key"`
	Text     string `json:"text"`
	Position int    `json:"position"`
}

func (v LibraryInput) Validate() error {
	if len(strings.TrimSpace(v.Name)) < 2 {
		return fmt.Errorf("library name is required")
	}
	return nil
}
func (v TermInput) Validate() error {
	if v.Concept == "" || len(v.Language) < 2 || v.Preferred == "" {
		return fmt.Errorf("term concept, language and preferred translation are required")
	}
	return nil
}
func (v DocumentInput) Validate() error {
	if v.ExternalID == "" || len(v.Language) < 2 || v.Fingerprint == "" || len(v.Fragments) == 0 {
		return fmt.Errorf("document identity, language, fingerprint and fragments are required")
	}
	for _, f := range v.Fragments {
		if f.Key == "" || strings.TrimSpace(f.Text) == "" || f.Position < 0 {
			return fmt.Errorf("fragment is invalid")
		}
	}
	return nil
}
func Now() time.Time { return time.Now().UTC() }
