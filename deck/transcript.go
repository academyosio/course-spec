package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Transcript is the part of a videoascode transcript.json that course-spec
// reads: each slide's narration per language. Other properties, such as the
// recording metadata, are ignored.
type Transcript struct {
	Slides []TranscriptSlide `json:"slides"`
}

// TranscriptSlide is one slide's narration in every language.
type TranscriptSlide struct {
	// ID is the slide's data-id. Older files have only Index.
	ID          string                        `json:"id,omitempty"`
	Index       int                           `json:"index"`
	Transcripts map[string]TranscriptLanguage `json:"transcripts"`
}

// TranscriptLanguage is a slide's narration in one language.
type TranscriptLanguage struct {
	Segments []TranscriptSegment `json:"segments"`
}

// TranscriptSegment is one sentence of narration.
type TranscriptSegment struct {
	Text string `json:"text"`
}

// Text joins the segments with single spaces, normalized.
func (l TranscriptLanguage) Text() string {
	parts := make([]string, 0, len(l.Segments))
	for _, s := range l.Segments {
		parts = append(parts, s.Text)
	}
	return Normalize(strings.Join(parts, " "))
}

// ReadTranscript reads a transcript file.
func ReadTranscript(path string) (*Transcript, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Transcript
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &t, nil
}

// Narration maps slide ID to language to normalized narration text.
type Narration map[string]map[string]string

// Match matches the transcript's slides to the deck's. Entries with an ID
// are matched by ID; when no entry has an ID, entries are matched by
// position and byPosition is true. Entries whose ID isn't in the deck are
// returned as unknown.
func (t *Transcript) Match(d *Deck) (n Narration, unknown []string, byPosition bool) {
	n = Narration{}
	anyID := false
	for _, s := range t.Slides {
		if s.ID != "" {
			anyID = true
			break
		}
	}
	put := func(id string, s TranscriptSlide) {
		m := n[id]
		if m == nil {
			m = map[string]string{}
			n[id] = m
		}
		for lang, content := range s.Transcripts {
			m[lang] = content.Text()
		}
	}
	if !anyID {
		for i, s := range t.Slides {
			if i < len(d.Slides) && d.Slides[i].ID != "" {
				put(d.Slides[i].ID, s)
			}
		}
		return n, nil, len(t.Slides) > 0
	}
	for _, s := range t.Slides {
		if s.ID == "" {
			continue
		}
		if _, ok := d.Slide(s.ID); !ok {
			unknown = append(unknown, s.ID)
			continue
		}
		put(s.ID, s)
	}
	return n, unknown, false
}
