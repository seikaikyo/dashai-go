package user

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// --- Database entity types ---

type User struct {
	ID               string         `json:"id"`
	AuthID           string         `json:"auth_id"`
	Email            *string        `json:"email"`
	DisplayName      *string        `json:"display_name"`
	BirthDate        *string        `json:"birth_date"` // ISO date string or null
	Plan             string         `json:"plan"`
	CreditsRemaining int            `json:"credits_remaining"`
	Preferences      map[string]any `json:"preferences"`
	HrCompany        map[string]any `json:"hr_company"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type Partner struct {
	ID               string   `json:"id"`
	UserID           string   `json:"user_id,omitempty"`
	Nickname         string   `json:"nickname"`
	BirthDate        string   `json:"birth_date"`
	Relation         string   `json:"relation"`
	SubTags          []string `json:"sub_tags"`
	EmotionDirection *string  `json:"emotion_direction"`
	CreatedAt        time.Time `json:"created_at"`
}

type Company struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id,omitempty"`
	Name        string  `json:"name"`
	FoundingDate *string `json:"founding_date"`
	Country     string  `json:"country"`
	Memo        *string `json:"memo"`
	JobURL      *string `json:"job_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type JobSeeker struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id,omitempty"`
	Name      string           `json:"name"`
	BirthDate string           `json:"birth_date"`
	Companies []map[string]any `json:"companies"`
	CreatedAt time.Time        `json:"created_at"`
}

type HrCandidate struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	BirthDate string    `json:"birth_date"`
	CreatedAt time.Time `json:"created_at"`
}

type CompanyCacheEntry struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	FoundingDate *string `json:"founding_date"`
	BusinessNo  *string `json:"business_no"`
	Source      *string `json:"source"`
	JobURL104   *string `json:"job_url_104"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- Request types ---

type ProfileSyncRequest struct {
	BirthDate   *string        `json:"birth_date"`
	DisplayName *string        `json:"display_name"`
	Preferences map[string]any `json:"preferences"`
}

type FullSyncRequest struct {
	BirthDate    *string           `json:"birth_date"`
	Preferences  map[string]any    `json:"preferences"`
	Partners     []PartnerData     `json:"partners"`
	Companies    []CompanyData     `json:"companies"`
	JobSeekers   []JobSeekerData   `json:"job_seekers"`
	HrCandidates []HrCandidateData `json:"hr_candidates"`
	HrCompany    map[string]any    `json:"hr_company"`
}

type PartnerData struct {
	ID               string   `json:"id"`
	Nickname         string   `json:"nickname"`
	BirthDate        string   `json:"birth_date"`
	Relation         string   `json:"relation"`
	SubTags          []string `json:"sub_tags"`
	EmotionDirection *string  `json:"emotion_direction"`
}

type CompanyData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	FoundingDate *string `json:"founding_date"`
	Country     string  `json:"country"`
	Memo        *string `json:"memo"`
	JobURL      *string `json:"job_url"`
}

type JobSeekerData struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	BirthDate string           `json:"birth_date"`
	Companies []map[string]any `json:"companies"`
}

type HrCandidateData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BirthDate string `json:"birth_date"`
}

type CompanyCacheBatchRequest struct {
	Names   []string `json:"names"`
	Country string   `json:"country"`
}

type CompanyCacheSaveRequest struct {
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	FoundingDate *string `json:"founding_date"`
	BusinessNo  *string `json:"business_no"`
	Source      *string `json:"source"`
	JobURL104   *string `json:"job_url_104"`
}

// --- Helpers ---

// newUUID generates a UUID v4 string using crypto/rand.
func newUUID() string {
	var uuid [16]byte
	rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// scannable is a common interface for pgx Row and Rows.
type scannable interface {
	Scan(dest ...any) error
}

// unmarshalJSON safely unmarshals JSON bytes, returning fallback on error.
func unmarshalJSON[T any](data []byte, fallback T) T {
	if len(data) == 0 {
		return fallback
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return fallback
	}
	return v
}

// datePtr formats a *time.Time as ISO date string pointer.
func datePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
