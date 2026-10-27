package coursespec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// DecodeError is a file that isn't valid for its type.
type DecodeError struct {
	File string
	// Pointer is a JSON pointer to the problem, when known.
	Pointer string
	Err     error
}

func (e *DecodeError) Error() string {
	if e.Pointer != "" {
		return fmt.Sprintf("%s: %s: %v", e.File, e.Pointer, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// Decode decodes JSON into v, rejecting unknown properties and trailing
// data. Every course-spec file goes through it, so another input syntax
// could be added by converting it to JSON first.
func Decode(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("unexpected data after the JSON value")
	}
	return nil
}

// ReadFile decodes the JSON file at path into v.
func ReadFile(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := Decode(data, v); err != nil {
		de := &DecodeError{File: path, Err: err}
		var te *json.UnmarshalTypeError
		if errors.As(err, &te) && te.Field != "" {
			de.Pointer = "/" + te.Field
		}
		return de
	}
	return nil
}

// ReadCourseManifest reads a course.json and checks its spec version.
func ReadCourseManifest(path string) (*CourseManifest, error) {
	var m CourseManifest
	if err := ReadFile(path, &m); err != nil {
		return nil, err
	}
	if err := CheckSpecVersion(m.SpecVersion); err != nil {
		return nil, &DecodeError{File: path, Pointer: "/specVersion", Err: err}
	}
	return &m, nil
}

// CheckSpecVersion accepts "1" and "1.<minor>", and refuses other majors.
func CheckSpecVersion(v string) error {
	if v == "" {
		return errors.New("specVersion is required")
	}
	major := v
	for i, r := range v {
		if r == '.' {
			major = v[:i]
			break
		}
	}
	if major != SpecVersion {
		return fmt.Errorf("specVersion %q is not supported; this reader supports version %s", v, SpecVersion)
	}
	return nil
}

// ReadSectionManifest reads a section.json.
func ReadSectionManifest(path string) (*SectionManifest, error) {
	var m SectionManifest
	if err := ReadFile(path, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ReadQuiz reads a quiz file.
func ReadQuiz(path string) (*Quiz, error) {
	var q Quiz
	if err := ReadFile(path, &q); err != nil {
		return nil, err
	}
	return &q, nil
}

// ReadPublished reads a published.json.
func ReadPublished(path string) (*Published, error) {
	var p Published
	if err := ReadFile(path, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ReadPronunciations reads a pronunciations.json.
func ReadPronunciations(path string) (Pronunciations, error) {
	var p Pronunciations
	if err := ReadFile(path, &p); err != nil {
		return nil, err
	}
	return p, nil
}
