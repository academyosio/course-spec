package coursespec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/academyosio/course-spec/deck"
)

// Course is a course directory resolved into one value.
type Course struct {
	// Dir is the course directory.
	Dir      string
	Manifest *CourseManifest
	Sections []Section
	// Pronunciations is nil when the course has none.
	Pronunciations Pronunciations
	// Published is nil when nothing has been published.
	Published *Published
}

// Section is a resolved section.
type Section struct {
	Ref SectionRef
	// Dir is the section directory.
	Dir string
	// Manifest is nil for an unreleased section without one.
	Manifest *SectionManifest
	// Deck is the parsed deck of an htmlDeck section.
	Deck *deck.Deck
	// Translations are the transcript's narration by slide and language.
	Translations deck.Narration
	Lessons      []Lesson
	// Quiz is nil when the section has none.
	Quiz *Quiz
}

// ID returns the section's ID, from its manifest or, for an unreleased
// section, its course manifest entry.
func (s Section) ID() string {
	if s.Manifest != nil {
		return s.Manifest.ID
	}
	return s.Ref.ID
}

// Title returns the section's title.
func (s Section) Title() string {
	if s.Manifest != nil {
		return s.Manifest.Title
	}
	return s.Ref.Title
}

// Load reads the course directory dir. It stops at the first problem; use
// the validate package to find every problem with a fix for each.
// Unreleased sections are loaded when their files are present and skipped
// otherwise.
func Load(dir string) (*Course, error) {
	m, err := ReadCourseManifest(filepath.Join(dir, CourseFile))
	if err != nil {
		return nil, err
	}
	c := &Course{Dir: dir, Manifest: m}

	if p, err := readOptional(filepath.Join(dir, PronunciationsFile), ReadPronunciations); err != nil {
		return nil, err
	} else if p != nil {
		c.Pronunciations = p
	}
	if p, err := readOptional(filepath.Join(dir, PublishedFile), ReadPublished); err != nil {
		return nil, err
	} else if p != nil {
		c.Published = p
	}

	for _, ref := range m.Sections {
		s, err := LoadSection(dir, ref, m)
		if err != nil {
			if !ref.IsReleased() {
				c.Sections = append(c.Sections, Section{Ref: ref, Dir: filepath.Join(dir, ref.Path)})
				continue
			}
			return nil, err
		}
		c.Sections = append(c.Sections, *s)
	}
	return c, nil
}

// LoadSection reads one section of the course in courseDir.
func LoadSection(courseDir string, ref SectionRef, cm *CourseManifest) (*Section, error) {
	sdir := filepath.Join(courseDir, ref.Path)
	sm, err := ReadSectionManifest(filepath.Join(sdir, SectionFile))
	if err != nil {
		return nil, err
	}
	s := &Section{Ref: ref, Dir: sdir, Manifest: sm}
	if sm.Source.Type == SourceHTMLDeck {
		if sm.Source.Path == "" {
			return nil, fmt.Errorf("%s: source.path is required for an htmlDeck source", filepath.Join(sdir, SectionFile))
		}
		deckPath := filepath.Join(sdir, sm.Source.Path)
		d, err := deck.ParseFile(deckPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", deckPath, err)
		}
		s.Deck = d
		tpath := filepath.Join(filepath.Dir(deckPath), TranscriptFile)
		t, err := readOptional(tpath, deck.ReadTranscript)
		if err != nil {
			return nil, err
		}
		if t != nil {
			n, unknown, _ := t.Match(d)
			if len(unknown) > 0 {
				return nil, fmt.Errorf("%s: slide %q is not in the deck", tpath, unknown[0])
			}
			s.Translations = n
		}
	}
	lessons, problems := ResolveLessons(sm, s.Deck)
	if len(problems) > 0 {
		return nil, fmt.Errorf("%s: %s", filepath.Join(sdir, SectionFile), problems[0].Message)
	}
	s.Lessons = lessons

	quizPath, err := QuizPath(sdir, sm)
	if err != nil {
		return nil, err
	}
	if quizPath != "" {
		q, err := ReadQuiz(quizPath)
		if err != nil {
			return nil, err
		}
		s.Quiz = q
	}
	return s, nil
}

// QuizPath returns the section's quiz file: the manifest's quiz, or
// quiz.json when it exists, or "" when the section has no quiz.
func QuizPath(sectionDir string, m *SectionManifest) (string, error) {
	if m.Quiz != "" {
		return filepath.Join(sectionDir, m.Quiz), nil
	}
	p := filepath.Join(sectionDir, QuizFile)
	_, err := os.Stat(p)
	switch {
	case err == nil:
		return p, nil
	case errors.Is(err, fs.ErrNotExist):
		return "", nil
	default:
		return "", err
	}
}

func readOptional[T any](path string, read func(string) (T, error)) (T, error) {
	var zero T
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return zero, nil
		}
		return zero, err
	}
	return read(path)
}

// Lesson returns the lesson with the given ID in any section.
func (c *Course) Lesson(id string) (Lesson, bool) {
	for _, s := range c.Sections {
		for _, l := range s.Lessons {
			if l.ID == id {
				return l, true
			}
		}
	}
	return Lesson{}, false
}

// Fingerprint returns a lesson's content fingerprint, or "" for a lesson
// without slides.
func (c *Course) Fingerprint(s Section, l Lesson) string {
	if len(l.Slides) == 0 {
		return ""
	}
	dl := deck.Lesson{ID: l.ID, Title: l.Title, Slides: l.Slides}
	return dl.Fingerprint(c.Manifest.AllLanguages(), c.Manifest.DefaultLanguage(), s.Translations)
}
