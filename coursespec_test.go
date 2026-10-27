package coursespec_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/internal/testcourse"
)

func TestLoad(t *testing.T) {
	c, err := coursespec.Load(testcourse.Source())
	if err != nil {
		t.Fatal(err)
	}
	if c.Manifest.ID != "building-web-apis" || len(c.Sections) != 3 {
		t.Fatalf("course %s with %d sections", c.Manifest.ID, len(c.Sections))
	}
	s := c.Sections[0]
	var ids []string
	for _, l := range s.Lessons {
		ids = append(ids, l.ID+":"+l.Type)
	}
	if got := strings.Join(ids, " "); got != "resources:video glossary:text methods:video" {
		t.Errorf("lessons = %s", got)
	}
	if !s.Lessons[0].Free || len(s.Lessons[0].Objectives) != 1 {
		t.Error("manifest details were not applied to the discovered lesson")
	}
	if got := s.Translations["methods"]["es-ES"]; got != "GET lee un recurso. PUT lo reemplaza." {
		t.Errorf("translation = %q", got)
	}
	if s.Quiz == nil || s.Quiz.ID != "designing-endpoints-quiz" {
		t.Error("section quiz not loaded")
	}
	final := c.Sections[1]
	if len(final.Lessons) != 0 || final.Quiz == nil {
		t.Error("quiz-only section not loaded")
	}
	planned := c.Sections[2]
	if planned.Ref.IsReleased() || planned.ID() != "versioning" || planned.Manifest != nil {
		t.Errorf("planned section = %+v", planned)
	}
	if c.Pronunciations["API"]["en-US"] != "A P I" {
		t.Error("pronunciations not loaded")
	}
	if _, ok := c.Lesson("methods"); !ok {
		t.Error("Lesson(methods) not found")
	}
}

func TestLoadUnknownProperty(t *testing.T) {
	dir := testcourse.Copy(t)
	testcourse.EditJSON(t, dir, "course.json", func(m map[string]any) { m["titel"] = "typo" })
	_, err := coursespec.Load(dir)
	var de *coursespec.DecodeError
	if !errors.As(err, &de) || !strings.Contains(err.Error(), "titel") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckSpecVersion(t *testing.T) {
	for v, ok := range map[string]bool{"1": true, "1.3": true, "2": false, "": false, "10": false} {
		if err := coursespec.CheckSpecVersion(v); (err == nil) != ok {
			t.Errorf("%q: err = %v", v, err)
		}
	}
}

func TestDecodeTrailingData(t *testing.T) {
	var v map[string]any
	if err := coursespec.Decode([]byte(`{} {}`), &v); err == nil {
		t.Fatal("expected an error for trailing data")
	}
}

func TestFind(t *testing.T) {
	dirs, err := coursespec.Find(filepath.Dir(testcourse.Source()))
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || filepath.Base(dirs[0]) != "web-apis" {
		t.Fatalf("dirs = %v", dirs)
	}
}

func TestResolveRetiredLesson(t *testing.T) {
	dir := testcourse.Copy(t)
	testcourse.EditJSON(t, dir, "sections/01-designing-endpoints/section.json", func(m map[string]any) {
		m["lessons"] = append(m["lessons"].([]any), map[string]any{"id": "old-lesson", "title": "Old", "retired": true})
	})
	c, err := coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	l, ok := c.Lesson("old-lesson")
	if !ok || !l.Retired {
		t.Fatalf("retired lesson = %+v %v", l, ok)
	}
}

func TestStatus(t *testing.T) {
	dir := testcourse.Copy(t)
	c, err := coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := c.Status()
	if r.Sections[0].Lessons[0].Status != coursespec.StatusNew || r.Sections[2].Released {
		t.Fatalf("status without published.json = %+v", r.Sections)
	}
	resources := r.Sections[0].Lessons[0]
	quiz := r.Sections[0].Quiz
	testcourse.Write(t, dir, "published.json", `{"courseVersion": "1.0.0",
		"lessons": {"resources": {"contentFingerprint": "`+resources.Fingerprint+`"},
		            "methods": {"contentFingerprint": "sha256:old"}},
		"quizzes": {"designing-endpoints-quiz": {"fingerprint": "`+quiz.Fingerprint+`"}}}`)
	c, err = coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	r = c.Status()
	got := map[string]string{}
	for _, l := range r.Sections[0].Lessons {
		got[l.ID] = l.Status
	}
	if got["resources"] != coursespec.StatusPublished || got["methods"] != coursespec.StatusChanged || got["glossary"] != coursespec.StatusNew {
		t.Errorf("statuses = %v", got)
	}
	if r.Sections[0].Quiz.Status != coursespec.StatusPublished {
		t.Errorf("quiz status = %s", r.Sections[0].Quiz.Status)
	}
}

func TestQuestionPointsDefault(t *testing.T) {
	two := 2
	if (coursespec.Question{}).PointsOrDefault() != 1 || (coursespec.Question{Points: &two}).PointsOrDefault() != 2 {
		t.Error("PointsOrDefault")
	}
}

func TestWritePublished(t *testing.T) {
	path := filepath.Join(t.TempDir(), "published.json")
	in := &coursespec.Published{Platforms: map[string]coursespec.PublishedPlatform{"academyos": {CourseID: "c1", Revision: 3}}}
	if err := coursespec.WritePublished(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := coursespec.ReadPublished(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.Platforms["academyos"].Revision != 3 {
		t.Errorf("revision = %d", out.Platforms["academyos"].Revision)
	}
}
