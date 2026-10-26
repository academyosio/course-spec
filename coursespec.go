// Package coursespec reads course directories in the course-spec format: a
// course manifest, a manifest per section, HTML decks with narration,
// quizzes, pronunciations, and the published state written by tools.
//
// Every file is JSON with camelCase property names. [Load] resolves a course
// directory into one [Course] value; the validate package checks one and
// reports what to fix.
package coursespec

// SpecVersion is the major version of the format this package reads.
const SpecVersion = "1"

// File names with fixed places in a course directory.
const (
	CourseFile         = "course.json"
	SectionFile        = "section.json"
	QuizFile           = "quiz.json"
	PublishedFile      = "published.json"
	PronunciationsFile = "pronunciations.json"
	TranscriptFile     = "voiceover/transcript.json"
	DefaultLanguage    = "en-US"
	DefaultTheme       = "theme"
)
