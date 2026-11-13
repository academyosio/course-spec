package deck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

// Recordings live in a directory beside the deck, like its transcript.
const (
	// RecordingsDir is the recordings directory, relative to the deck's
	// directory.
	RecordingsDir = "recordings"
	// RecordingsFile is the recordings index, relative to the deck's
	// directory.
	RecordingsFile = "recordings/recordings.json"
)

// Recording statuses.
const (
	RecordingCurrent = "current" // recorded from the lesson's narration as it is
	RecordingStale   = "stale"   // the narration changed after recording
	RecordingMissing = "missing" // not recorded in that language
)

// Recordings is a deck's recorded narration: recordings/recordings.json.
type Recordings struct {
	// Lessons has one entry per lesson and language.
	Lessons []LessonRecording `json:"lessons"`
}

// LessonRecording is one lesson's narration recorded in one language.
type LessonRecording struct {
	// Lesson is the lesson's ID: its first slide's data-id.
	Lesson string `json:"lesson" jsonschema:"minLength=1"`
	// Language is the narration's BCP-47 language.
	Language string `json:"language" jsonschema:"minLength=1"`
	// Audio is the MP3 file, relative to the recordings directory.
	Audio string `json:"audio" jsonschema:"minLength=1"`
	// Captions is a WebVTT file, relative to the recordings directory.
	Captions string `json:"captions,omitempty"`
	// DurationMs is the audio's length.
	DurationMs int `json:"durationMs" jsonschema:"minimum=0"`
	// Voice is the voice it was recorded with.
	Voice *RecordingVoice `json:"voice,omitempty"`
	// Slides are the lesson's slides in order, with where each one's
	// narration starts in the audio.
	Slides []RecordedSlide `json:"slides" jsonschema:"minItems=1"`
}

// RecordingVoice describes the voice a recording was made with.
type RecordingVoice struct {
	// Provider is the TTS provider, such as elevenlabs.
	Provider string `json:"provider,omitempty"`
	// VoiceID is the provider's voice ID.
	VoiceID string `json:"voiceId,omitempty"`
	// Name is the voice's display name.
	Name string `json:"name,omitempty"`
	// Model is the provider's model.
	Model string `json:"model,omitempty"`
}

// RecordedSlide is one slide's place in a lesson recording.
type RecordedSlide struct {
	// ID is the slide's data-id.
	ID string `json:"id"`
	// StartMs is where the slide's narration starts in the audio.
	StartMs int `json:"startMs" jsonschema:"minimum=0"`
	// DurationMs covers the slide's speech and its pause.
	DurationMs int `json:"durationMs" jsonschema:"minimum=0"`
	// Narration is the NarrationFingerprint of the narration recorded.
	Narration string `json:"narration"`
}

// ReadRecordings reads a recordings file.
func ReadRecordings(p string) (*Recordings, error) {
	data, err := os.ReadFile(p) //nolint:gosec // a path the caller chose
	if err != nil {
		return nil, err
	}
	var r Recordings
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("%s: %w", p, err)
	}
	return &r, nil
}

// Write writes the recordings as indented JSON with a trailing newline,
// sorted by lesson then language, so rewriting it changes only what
// changed.
func (r *Recordings) Write(p string) error {
	slices.SortStableFunc(r.Lessons, func(a, b LessonRecording) int {
		if c := strings.Compare(a.Lesson, b.Lesson); c != 0 {
			return c
		}
		return strings.Compare(a.Language, b.Language)
	})
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644) //nolint:gosec // course files are meant to be read
}

// Find returns the lesson's recording in a language, or nil.
func (r *Recordings) Find(lesson, lang string) *LessonRecording {
	if r == nil {
		return nil
	}
	for i := range r.Lessons {
		if r.Lessons[i].Lesson == lesson && r.Lessons[i].Language == lang {
			return &r.Lessons[i]
		}
	}
	return nil
}

// Set adds a recording, replacing any for the same lesson and language.
func (r *Recordings) Set(rec LessonRecording) {
	if old := r.Find(rec.Lesson, rec.Language); old != nil {
		*old = rec
		return
	}
	r.Lessons = append(r.Lessons, rec)
}

// CheckPaths reports a problem with the recording's file names or timings:
// paths must be clean, relative and inside the recordings directory, the
// audio MP3 and the captions WebVTT, and the slides must tile the audio.
func (rec *LessonRecording) CheckPaths() error {
	if err := checkRecordingPath(rec.Audio, ".mp3"); err != nil {
		return fmt.Errorf("audio: %w", err)
	}
	if rec.Captions != "" {
		if err := checkRecordingPath(rec.Captions, ".vtt"); err != nil {
			return fmt.Errorf("captions: %w", err)
		}
	}
	if len(rec.Slides) == 0 {
		return fmt.Errorf("no slides")
	}
	at := 0
	for _, s := range rec.Slides {
		if s.ID == "" {
			return fmt.Errorf("a slide has no ID")
		}
		if s.StartMs != at || s.DurationMs < 0 {
			return fmt.Errorf("slide %s starts at %d ms; the slides before it end at %d ms", s.ID, s.StartMs, at)
		}
		at += s.DurationMs
	}
	if rec.DurationMs < at {
		return fmt.Errorf("the slides last %d ms, longer than the audio's %d ms", at, rec.DurationMs)
	}
	return nil
}

func checkRecordingPath(p, ext string) error {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || path.Clean(p) != p ||
		p == ".." || strings.HasPrefix(p, "../") || (len(p) >= 2 && p[1] == ':') {
		return fmt.Errorf("%q isn't a path inside the recordings directory", p)
	}
	if !strings.EqualFold(path.Ext(p), ext) {
		return fmt.Errorf("%q must be a %s file", p, ext)
	}
	return nil
}

// NarrationFingerprint identifies a slide's narration in one language: its
// normalized text and its total pause. A recording stores it per slide, so
// a player can tell which slides changed after recording.
func NarrationFingerprint(text string, pauseMs int) string {
	h := sha256.New()
	writeField(h, []byte(Normalize(text)))
	writeField(h, []byte(strconv.Itoa(pauseMs)))
	return FingerprintPrefix + hex.EncodeToString(h.Sum(nil))
}

// SlideNarration returns the narration fingerprint of each of the lesson's
// slides in lang: the deck's own narration and pause for defaultLang, the
// transcript's translation with no pause for other languages.
func (l Lesson) SlideNarration(lang, defaultLang string, translations Narration) []string {
	fps := make([]string, len(l.Slides))
	for i, s := range l.Slides {
		text, pause := s.Narration, s.PauseMs
		if lang != defaultLang {
			text, pause = translations[s.ID][lang], 0
		}
		fps[i] = NarrationFingerprint(text, pause)
	}
	return fps
}

// RecordingCheck compares a recording with its lesson.
type RecordingCheck struct {
	// Status is RecordingCurrent, RecordingStale or RecordingMissing.
	Status string `json:"status"`
	// StaleSlides are the lesson's slides whose narration differs from
	// what was recorded, or that weren't recorded.
	StaleSlides []string `json:"staleSlides,omitempty"`
	// Removed are recorded slides no longer in the lesson.
	Removed []string `json:"removed,omitempty"`
}

// CheckRecording compares rec, the lesson's recording in rec.Language, with
// the lesson's narration. It is current only when the recorded slides are
// the lesson's, in order, each recorded from its narration as it is. Slide
// HTML isn't compared, so restyling a slide keeps its recording. A nil rec
// is missing.
func (l Lesson) CheckRecording(rec *LessonRecording, defaultLang string, translations Narration) RecordingCheck {
	if rec == nil {
		return RecordingCheck{Status: RecordingMissing}
	}
	fps := l.SlideNarration(rec.Language, defaultLang, translations)
	recorded := map[string]string{}
	for _, s := range rec.Slides {
		recorded[s.ID] = s.Narration
	}
	var c RecordingCheck
	inLesson := map[string]bool{}
	for i, s := range l.Slides {
		inLesson[s.ID] = true
		if fp, ok := recorded[s.ID]; !ok || fp != fps[i] {
			c.StaleSlides = append(c.StaleSlides, s.ID)
		}
	}
	for _, s := range rec.Slides {
		if !inLesson[s.ID] {
			c.Removed = append(c.Removed, s.ID)
		}
	}
	inOrder := len(rec.Slides) == len(l.Slides)
	for i := 0; inOrder && i < len(l.Slides); i++ {
		inOrder = rec.Slides[i].ID == l.Slides[i].ID
	}
	c.Status = RecordingCurrent
	if len(c.StaleSlides) > 0 || len(c.Removed) > 0 || !inOrder {
		c.Status = RecordingStale
	}
	return c
}
