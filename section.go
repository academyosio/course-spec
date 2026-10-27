package coursespec

// Source types.
const (
	SourceHTMLDeck = "htmlDeck"
	SourceMarp     = "marp"
	SourceVideo    = "video"
	SourceNone     = "none"
)

// Lesson types.
const (
	LessonVideo = "video"
	LessonText  = "text"
)

// SectionManifest is a section's section.json.
type SectionManifest struct {
	ID      string `json:"id" jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	Source  Source `json:"source"`
	// Lessons adds details to discovered lessons by ID, or lists lessons for
	// sources that don't define them.
	Lessons []LessonEntry `json:"lessons,omitempty"`
	// Quiz is the quiz file; it defaults to quiz.json when that exists.
	Quiz string `json:"quiz,omitempty"`
}

// Source is where a section's lessons and video come from.
type Source struct {
	// Type is htmlDeck, marp, video, or none.
	Type string `json:"type" jsonschema:"enum=htmlDeck,enum=marp,enum=video,enum=none"`
	// Path is the deck file, relative to the section directory.
	Path string `json:"path,omitempty"`
}

// LessonEntry describes a lesson in a section manifest.
type LessonEntry struct {
	ID string `json:"id" jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"`
	// Title is required when the lesson isn't discovered from a deck, and
	// overrides a discovered title.
	Title string `json:"title,omitempty"`
	// Type is video or text.
	Type string `json:"type,omitempty" jsonschema:"enum=video,enum=text"`
	// Content is a Markdown file for a text lesson.
	Content          string   `json:"content,omitempty"`
	Objectives       []string `json:"objectives,omitempty"`
	Free             bool     `json:"free,omitempty"`
	EstimatedMinutes int      `json:"estimatedMinutes,omitempty" jsonschema:"minimum=0"`
	// After places a text lesson after the named lesson.
	After string `json:"after,omitempty"`
	// File is a video file for a video source, not committed.
	File string `json:"file,omitempty"`
	// Retired keeps a published ID that new learners don't see.
	Retired bool `json:"retired,omitempty"`
}
