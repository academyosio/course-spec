package coursespec

import (
	"fmt"

	"github.com/academyosio/course-spec/deck"
)

// Lesson is a resolved lesson: discovered from a deck or listed in the
// section manifest, with the manifest's details applied.
type Lesson struct {
	ID    string
	Title string
	// Type is video or text.
	Type             string
	Content          string
	Objectives       []string
	Free             bool
	EstimatedMinutes int
	Retired          bool
	// Slides are the deck slides of a lesson discovered from a deck.
	Slides []deck.Slide
	// Discovered reports whether the lesson came from the deck.
	Discovered bool
}

// LessonProblem is a section manifest entry that couldn't be resolved.
type LessonProblem struct {
	// Index is the entry's position in the manifest's lessons.
	Index   int
	Message string
	Fix     string
}

// ResolveLessons combines a section's discovered and listed lessons. d is
// the parsed deck for an htmlDeck section and nil otherwise.
func ResolveLessons(m *SectionManifest, d *deck.Deck) ([]Lesson, []LessonProblem) {
	var problems []LessonProblem
	if m.Source.Type != SourceHTMLDeck || d == nil {
		lessons := make([]Lesson, 0, len(m.Lessons))
		for _, e := range m.Lessons {
			l := fromEntry(e)
			if l.Type == "" {
				l.Type = LessonText
				if m.Source.Type == SourceVideo {
					l.Type = LessonVideo
				}
			}
			lessons = append(lessons, l)
		}
		return lessons, problems
	}

	var lessons []Lesson
	index := map[string]int{}
	for _, dl := range d.Lessons(m.ID) {
		index[dl.ID] = len(lessons)
		lessons = append(lessons, Lesson{ID: dl.ID, Title: dl.Title, Type: LessonVideo, Slides: dl.Slides, Discovered: true})
	}
	var texts []struct {
		lesson Lesson
		after  string
	}
	for i, e := range m.Lessons {
		if e.Type == LessonText {
			texts = append(texts, struct {
				lesson Lesson
				after  string
			}{fromEntry(e), e.After})
			continue
		}
		j, ok := index[e.ID]
		if !ok && e.Retired {
			l := fromEntry(e)
			l.Type = LessonVideo
			lessons = append(lessons, l)
			continue
		}
		if !ok {
			problems = append(problems, LessonProblem{
				Index:   i,
				Message: fmt.Sprintf("lesson %q is not in the deck", e.ID),
				Fix:     "use the data-id of a slide with data-lesson, or set \"type\": \"text\" for a text lesson",
			})
			continue
		}
		applyEntry(&lessons[j], e)
	}
	for _, t := range texts {
		pos := len(lessons)
		if t.after != "" {
			for k, l := range lessons {
				if l.ID == t.after {
					pos = k + 1
				}
			}
		}
		lessons = append(lessons[:pos], append([]Lesson{t.lesson}, lessons[pos:]...)...)
	}
	return lessons, problems
}

func fromEntry(e LessonEntry) Lesson {
	l := Lesson{ID: e.ID, Title: e.Title, Type: e.Type}
	applyEntry(&l, e)
	return l
}

func applyEntry(l *Lesson, e LessonEntry) {
	if e.Title != "" {
		l.Title = e.Title
	}
	if e.Content != "" {
		l.Content = e.Content
	}
	l.Objectives = e.Objectives
	l.Free = e.Free
	l.EstimatedMinutes = e.EstimatedMinutes
	l.Retired = e.Retired
}
