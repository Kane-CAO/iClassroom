package domain

import "time"

// DisplayMode controls how a featured answer is shown on the big screen.
type DisplayMode string

const (
	// DisplayAnonymous hides the student/group identity.
	DisplayAnonymous DisplayMode = "anonymous"
	// DisplayShowGroup reveals the owning group.
	DisplayShowGroup DisplayMode = "showGroup"
)

// Valid reports whether m is a known display mode.
func (m DisplayMode) Valid() bool {
	switch m {
	case DisplayAnonymous, DisplayShowGroup:
		return true
	default:
		return false
	}
}

// FeaturedAnswer marks a submission as a highlighted answer for the big screen.
// A submission can be featured at most once (enforced by a DB unique constraint
// on submission_id).
type FeaturedAnswer struct {
	ID           int64       `json:"featuredId"`
	RoomID       int64       `json:"-"`
	SubmissionID int64       `json:"submissionId"`
	DisplayMode  DisplayMode `json:"displayMode"`
	CreatedAt    time.Time   `json:"createdAt"`
}
