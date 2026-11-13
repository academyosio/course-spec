package validate_test

import (
	"path/filepath"
	"testing"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/deck"
	"github.com/academyosio/course-spec/internal/testcourse"
	"github.com/academyosio/course-spec/validate"
)

const recRel = sec1 + "deck/" + deck.RecordingsFile

// recordFirstLesson writes a current recording of the first lesson in both
// course languages and returns the recordings.
func recordFirstLesson(t *testing.T, dir string) *deck.Recordings {
	t.Helper()
	c, err := coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := c.Sections[0]
	l := s.Lessons[0]
	dl := deck.Lesson{ID: l.ID, Slides: l.Slides}
	r := &deck.Recordings{}
	for _, lang := range c.Manifest.AllLanguages() {
		rec := deck.LessonRecording{Lesson: l.ID, Language: lang, Audio: lang + "/" + l.ID + ".mp3", Captions: lang + "/" + l.ID + ".vtt"}
		for i, fp := range dl.SlideNarration(lang, c.Manifest.DefaultLanguage(), s.Translations) {
			rec.Slides = append(rec.Slides, deck.RecordedSlide{ID: l.Slides[i].ID, StartMs: i * 1000, DurationMs: 1000, Narration: fp})
		}
		rec.DurationMs = len(rec.Slides) * 1000
		r.Set(rec)
		testcourse.Write(t, dir, sec1+"deck/recordings/"+rec.Audio, "mp3")
		testcourse.Write(t, dir, sec1+"deck/recordings/"+rec.Captions, "WEBVTT\n")
	}
	if err := r.Write(filepath.Join(dir, filepath.FromSlash(recRel))); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRecordings(t *testing.T) {
	dir := testcourse.Copy(t)
	recordFirstLesson(t, dir)
	if fs := validate.Course(dir, validate.Options{}); len(fs) != 0 {
		t.Fatalf("findings for a current recording:\n%s", list(fs))
	}

	c, err := coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	lr := c.Status().Sections[0].Lessons
	if lr[0].Recordings["en-US"] != deck.RecordingCurrent || lr[0].Recordings["es-ES"] != deck.RecordingCurrent {
		t.Errorf("first lesson = %v", lr[0].Recordings)
	}
	others := 0
	for _, l := range lr[1:] {
		want := deck.RecordingMissing
		if l.Slides == 0 {
			want = "" // a text lesson has no recording status
		}
		if l.Recordings["en-US"] != want {
			t.Errorf("lesson %s = %v", l.ID, l.Recordings)
		}
		if l.Slides > 0 {
			others++
		}
	}
	if others == 0 {
		t.Error("the sample course needs a second deck lesson")
	}
}

func TestRecordingRules(t *testing.T) {
	cases := []struct {
		name, severity, code string
		edit                 func(r *deck.Recordings, dir string)
	}{
		{"unknown lesson", validate.Error, "unknown-lesson", func(r *deck.Recordings, _ string) { r.Lessons[0].Lesson = "gone" }},
		{"unknown language", validate.Error, "recording-language", func(r *deck.Recordings, _ string) { r.Lessons[0].Language = "fr-FR" }},
		{"duplicate", validate.Error, "duplicate-recording", func(r *deck.Recordings, _ string) { r.Lessons = append(r.Lessons, r.Lessons[0]) }},
		{"path outside", validate.Error, "invalid-recording", func(r *deck.Recordings, _ string) { r.Lessons[0].Audio = "../x.mp3" }},
		{"timings", validate.Error, "invalid-recording", func(r *deck.Recordings, _ string) { r.Lessons[0].Slides[0].DurationMs = 5 }},
		{"missing audio", validate.Error, "missing-file", func(r *deck.Recordings, _ string) { r.Lessons[0].Audio = "en-US/none.mp3" }},
		{"stale", validate.Warning, "stale-recording", func(r *deck.Recordings, _ string) { r.Lessons[0].Slides[0].Narration = "sha256:old" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := testcourse.Copy(t)
			r := recordFirstLesson(t, dir)
			tc.edit(r, dir)
			if err := r.Write(filepath.Join(dir, filepath.FromSlash(recRel))); err != nil {
				t.Fatal(err)
			}
			expect(t, validate.Course(dir, validate.Options{}), tc.severity, tc.code)
		})
	}
	t.Run("invalid json", func(t *testing.T) {
		dir := testcourse.Copy(t)
		testcourse.Write(t, dir, recRel, "{")
		expect(t, validate.Course(dir, validate.Options{}), validate.Error, "invalid-json")
	})
}
