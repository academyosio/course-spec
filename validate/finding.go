// Package validate checks a course directory against the course-spec rules
// and reports each problem with its file, location, and a fix.
package validate

import "fmt"

// Severity of a finding.
const (
	Error   = "error"
	Warning = "warning"
)

// Finding is one problem in a course.
type Finding struct {
	// Severity is error or warning. Errors make a course invalid.
	Severity string `json:"severity"`
	// Code names the rule, such as "lesson-not-in-deck".
	Code string `json:"code"`
	// File is relative to the course directory, with forward slashes.
	File string `json:"file"`
	// Pointer is a JSON pointer into File, or a slide reference for decks.
	Pointer string `json:"pointer,omitempty"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}

func (f Finding) String() string {
	loc := f.File
	if f.Pointer != "" {
		loc += " " + f.Pointer
	}
	s := fmt.Sprintf("%s: %s: %s [%s]", f.Severity, loc, f.Message, f.Code)
	if f.Fix != "" {
		s += "\n  fix: " + f.Fix
	}
	return s
}

// Options adjusts the checks.
type Options struct {
	// StrictNarration warns about characters TTS reads badly.
	StrictNarration bool
}

// Counts returns the number of errors and warnings.
func Counts(fs []Finding) (errors, warnings int) {
	for _, f := range fs {
		if f.Severity == Error {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
}
