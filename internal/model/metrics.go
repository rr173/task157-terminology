package model

import "time"

type LanguageCoverage struct {
	Language      string  `json:"language"`
	Documents     int     `json:"documents"`
	Fragments     int     `json:"fragments"`
	Hits          int     `json:"hits"`
	Problematic   int     `json:"problematic"`
	CoverageRatio float64 `json:"coverage_ratio"`
}

type CoverageReport struct {
	LibraryID       string             `json:"library_id"`
	Documents       int                `json:"documents"`
	Fragments       int                `json:"fragments"`
	Hits            int                `json:"hits"`
	ProblematicHits int                `json:"problematic_hits"`
	Suggestions     int                `json:"suggestions"`
	Languages       []LanguageCoverage `json:"languages"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

type TermUsage struct {
	Concept       string `json:"concept"`
	Language      string `json:"language"`
	Preferred     string `json:"preferred"`
	Occurrences   int    `json:"occurrences"`
	ForbiddenHits int    `json:"forbidden_hits"`
	Documents     int    `json:"documents"`
}

type ExportSnapshot struct {
	Library     Library       `json:"library"`
	Terms       []Term        `json:"terms"`
	Documents   []Document    `json:"documents"`
	Tasks       []CheckTask   `json:"tasks"`
	Suggestions []Suggestion  `json:"suggestions"`
	Audit       []AuditRecord `json:"audit"`
	ExportedAt  time.Time     `json:"exported_at"`
}

type HealthReport struct {
	Healthy      bool      `json:"healthy"`
	Libraries    int       `json:"libraries"`
	Documents    int       `json:"documents"`
	RunningTasks int       `json:"running_tasks"`
	OpenReviews  int       `json:"open_reviews"`
	CheckedAt    time.Time `json:"checked_at"`
}
