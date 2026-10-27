package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/academyosio/course-spec/internal/testcourse"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCommand(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestValidateValid(t *testing.T) {
	out, err := run(t, "validate", testcourse.Source())
	if err != nil || !strings.Contains(out, "0 errors, 0 warnings") {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	dir := testcourse.Copy(t)
	testcourse.EditJSON(t, dir, "course.json", func(m map[string]any) { m["id"] = "Bad ID" })
	out, err := run(t, "validate", "--format", "json", dir)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("err = %v", err)
	}
	var v struct {
		Courses []struct {
			Errors   int `json:"errors"`
			Findings []struct {
				Code string `json:"code"`
			} `json:"findings"`
		} `json:"courses"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if len(v.Courses) != 1 || v.Courses[0].Errors != 1 || v.Courses[0].Findings[0].Code != "invalid-id" {
		t.Errorf("result = %+v", v)
	}
}

func TestValidateFindsCourses(t *testing.T) {
	out, err := run(t, "validate", "../../testdata")
	if err != nil || !strings.Contains(out, "web-apis: 0 errors") {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	if _, err := run(t, "validate", t.TempDir()); err == nil {
		t.Error("expected an error for a directory with no course")
	}
}

func TestLessons(t *testing.T) {
	out, err := run(t, "lessons", testcourse.Source())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"building-web-apis 1.0.0", "new        resources: Resources (2 slides)", "glossary: Glossary (text)", "final-assessment-quiz", "[planned]"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	out, err = run(t, "lessons", "--format", "json", testcourse.Source())
	if err != nil || !strings.Contains(out, `"status": "new"`) {
		t.Fatalf("json out = %q, err = %v", out, err)
	}
}

func TestLessonsInvalidCourse(t *testing.T) {
	dir := testcourse.Copy(t)
	testcourse.Remove(t, dir, "sections/01-designing-endpoints/deck/index.html")
	if _, err := run(t, "lessons", dir); err == nil || !strings.Contains(err.Error(), "coursespec validate") {
		t.Fatalf("err = %v", err)
	}
}

func TestSchema(t *testing.T) {
	out, err := run(t, "schema")
	if err != nil || out != "course\npublished\nquiz\nsection\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	out, err = run(t, "schema", "quiz")
	if err != nil || !strings.Contains(out, `"multipleChoice"`) {
		t.Fatalf("err = %v", err)
	}
	if _, err := run(t, "schema", "nope"); err == nil {
		t.Error("expected an error for an unknown schema")
	}
}

func TestBadFormat(t *testing.T) {
	if _, err := run(t, "validate", "--format", "yaml", testcourse.Source()); err == nil {
		t.Error("expected an error for an unknown format")
	}
}
