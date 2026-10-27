package coursespec

import (
	"encoding/json"
	"fmt"
	"os"
)

// Published is a course's published.json, written by tools after
// recording, uploading, or syncing. Keys are course-spec IDs.
type Published struct {
	CourseVersion string                       `json:"courseVersion,omitempty"`
	Lessons       map[string]PublishedLesson   `json:"lessons,omitempty"`
	Quizzes       map[string]PublishedQuiz     `json:"quizzes,omitempty"`
	Video         *PublishedVideo              `json:"video,omitempty"`
	Platforms     map[string]PublishedPlatform `json:"platforms,omitempty"`
}

// PublishedLesson records what was published for one lesson.
type PublishedLesson struct {
	// ContentFingerprint is the content fingerprint the video was made from.
	ContentFingerprint string `json:"contentFingerprint,omitempty"`
	// RecordingFingerprint is the recording tool's own, opaque fingerprint.
	RecordingFingerprint string                 `json:"recordingFingerprint,omitempty"`
	DurationMs           int                    `json:"durationMs,omitempty" jsonschema:"minimum=0"`
	Videos               map[string]VideoRef    `json:"videos,omitempty"`
	Captions             map[string]CaptionsRef `json:"captions,omitempty"`
}

// VideoRef names an uploaded video.
type VideoRef struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
}

// CaptionsRef records where captions were published.
type CaptionsRef struct {
	Provider string `json:"provider"`
}

// PublishedQuiz records a published quiz.
type PublishedQuiz struct {
	Fingerprint string `json:"fingerprint,omitempty"`
}

// PublishedVideo records course-level video publishing.
type PublishedVideo struct {
	Playlists map[string]VideoRef `json:"playlists,omitempty"`
}

// PublishedPlatform records one platform's IDs for the course.
type PublishedPlatform struct {
	BaseURL  string            `json:"baseUrl,omitempty"`
	CourseID string            `json:"courseId,omitempty"`
	Lessons  map[string]string `json:"lessons,omitempty"`
	Quizzes  map[string]string `json:"quizzes,omitempty"`
	// Revision is the platform's revision of the course when this copy was
	// last synced, so the platform can refuse a push from a stale copy.
	Revision int `json:"revision,omitempty" jsonschema:"minimum=0"`
}

// WritePublished writes p to path as indented JSON, the form tools commit.
func WritePublished(path string, p *Published) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("encode published state: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
