package deck

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// recordingDeck has one lesson of three slides; the second has a pause.
const recordingDeck = `<html><body>
<section class="slide" data-id="a" data-lesson="One" data-vo="First slide."></section>
<section class="slide" data-id="b" data-vo="Second slide. [PAUSE:800]"></section>
<section class="slide" data-id="c" data-vo="Third   slide."></section>
</body></html>`

func recordedLesson(t *testing.T) (Lesson, *LessonRecording) {
	t.Helper()
	l := mustParse(t, recordingDeck).Lessons("s")[0]
	fps := l.SlideNarration("en-US", "en-US", nil)
	rec := &LessonRecording{Lesson: "a", Language: "en-US", Audio: "en-US/a.mp3", DurationMs: 6000}
	at := 0
	for i, s := range l.Slides {
		rec.Slides = append(rec.Slides, RecordedSlide{ID: s.ID, StartMs: at, DurationMs: 2000, Narration: fps[i]})
		at += 2000
	}
	return l, rec
}

func TestNarrationFingerprint(t *testing.T) {
	a := NarrationFingerprint("Hello  there.", 500)
	if a != NarrationFingerprint("Hello there.", 500) {
		t.Error("whitespace changed the fingerprint")
	}
	if a == NarrationFingerprint("Hello there.", 0) || a == NarrationFingerprint("Hello there!", 500) {
		t.Error("a different pause or text kept the fingerprint")
	}
	if !strings.HasPrefix(a, FingerprintPrefix) {
		t.Errorf("fingerprint = %s", a)
	}
}

func TestCheckRecording(t *testing.T) {
	l, rec := recordedLesson(t)
	if c := l.CheckRecording(rec, "en-US", nil); c.Status != RecordingCurrent || c.StaleSlides != nil {
		t.Fatalf("fresh recording: %+v", c)
	}
	if c := l.CheckRecording(nil, "en-US", nil); c.Status != RecordingMissing {
		t.Errorf("nil recording: %+v", c)
	}

	edited := l
	edited.Slides = append([]Slide(nil), l.Slides...)
	edited.Slides[1].Narration = "Second slide, edited."
	if c := edited.CheckRecording(rec, "en-US", nil); c.Status != RecordingStale || !reflect.DeepEqual(c.StaleSlides, []string{"b"}) {
		t.Errorf("edited narration: %+v", c)
	}

	repaused := l
	repaused.Slides = append([]Slide(nil), l.Slides...)
	repaused.Slides[1].PauseMs = 100
	if c := repaused.CheckRecording(rec, "en-US", nil); !reflect.DeepEqual(c.StaleSlides, []string{"b"}) {
		t.Errorf("changed pause: %+v", c)
	}

	restyled := l
	restyled.Slides = append([]Slide(nil), l.Slides...)
	restyled.Slides[0].HTML = []byte("<section>new look</section>")
	if c := restyled.CheckRecording(rec, "en-US", nil); c.Status != RecordingCurrent {
		t.Errorf("restyled slide: %+v", c)
	}

	added := l
	added.Slides = append(append([]Slide(nil), l.Slides...), Slide{ID: "d", Narration: "Fourth."})
	if c := added.CheckRecording(rec, "en-US", nil); c.Status != RecordingStale || !reflect.DeepEqual(c.StaleSlides, []string{"d"}) {
		t.Errorf("added slide: %+v", c)
	}

	removed := l
	removed.Slides = l.Slides[:2]
	if c := removed.CheckRecording(rec, "en-US", nil); c.Status != RecordingStale || !reflect.DeepEqual(c.Removed, []string{"c"}) {
		t.Errorf("removed slide: %+v", c)
	}

	reordered := l
	reordered.Slides = []Slide{l.Slides[0], l.Slides[2], l.Slides[1]}
	if c := reordered.CheckRecording(rec, "en-US", nil); c.Status != RecordingStale || c.StaleSlides != nil {
		t.Errorf("reordered slides: %+v", c)
	}
}

func TestCheckRecordingTranslation(t *testing.T) {
	l, _ := recordedLesson(t)
	tr := Narration{"a": {"it-IT": "Prima."}, "b": {"it-IT": "Seconda."}, "c": {"it-IT": "Terza."}}
	fps := l.SlideNarration("it-IT", "en-US", tr)
	rec := &LessonRecording{Lesson: "a", Language: "it-IT"}
	for i, s := range l.Slides {
		rec.Slides = append(rec.Slides, RecordedSlide{ID: s.ID, Narration: fps[i]})
	}
	if c := l.CheckRecording(rec, "en-US", tr); c.Status != RecordingCurrent {
		t.Errorf("translation: %+v", c)
	}
	tr["b"]["it-IT"] = "Seconda, cambiata."
	if c := l.CheckRecording(rec, "en-US", tr); !reflect.DeepEqual(c.StaleSlides, []string{"b"}) {
		t.Errorf("edited translation: %+v", c)
	}
	if fps[1] != NarrationFingerprint("Seconda.", 0) {
		t.Error("a translation's fingerprint should have no pause")
	}
}

func TestRecordingsReadWrite(t *testing.T) {
	_, rec := recordedLesson(t)
	r := &Recordings{}
	r.Set(LessonRecording{Lesson: "z", Language: "en-US", Audio: "en-US/z.mp3", Slides: []RecordedSlide{{ID: "z"}}})
	r.Set(*rec)
	it := *rec
	it.Language, it.Audio = "it-IT", "it-IT/a.mp3"
	r.Set(it)
	rec.DurationMs = 7000
	r.Set(*rec) // replaces
	if len(r.Lessons) != 3 {
		t.Fatalf("lessons = %d", len(r.Lessons))
	}
	p := filepath.Join(t.TempDir(), "recordings.json")
	if err := r.Write(p); err != nil {
		t.Fatal(err)
	}
	got, err := ReadRecordings(p)
	if err != nil {
		t.Fatal(err)
	}
	order := []string{got.Lessons[0].Lesson + "/" + got.Lessons[0].Language, got.Lessons[1].Lesson + "/" + got.Lessons[1].Language, got.Lessons[2].Lesson}
	if !reflect.DeepEqual(order, []string{"a/en-US", "a/it-IT", "z"}) {
		t.Errorf("order = %v", order)
	}
	if f := got.Find("a", "en-US"); f == nil || f.DurationMs != 7000 || len(f.Slides) != 3 {
		t.Errorf("find = %+v", f)
	}
	if got.Find("a", "fr-FR") != nil || (*Recordings)(nil).Find("a", "en-US") != nil {
		t.Error("found a recording that isn't there")
	}
}

func TestCheckPaths(t *testing.T) {
	_, rec := recordedLesson(t)
	rec.Captions = "en-US/a.vtt"
	if err := rec.CheckPaths(); err != nil {
		t.Fatal(err)
	}
	for name, mod := range map[string]func(*LessonRecording){
		"outside":         func(r *LessonRecording) { r.Audio = "../a.mp3" },
		"absolute":        func(r *LessonRecording) { r.Audio = "/a.mp3" },
		"not clean":       func(r *LessonRecording) { r.Audio = "en-US/./a.mp3" },
		"backslash":       func(r *LessonRecording) { r.Audio = `en-US\a.mp3` },
		"not mp3":         func(r *LessonRecording) { r.Audio = "en-US/a.wav" },
		"captions":        func(r *LessonRecording) { r.Captions = "en-US/a.srt" },
		"gap":             func(r *LessonRecording) { r.Slides[1].StartMs = 2500 },
		"longer than mp3": func(r *LessonRecording) { r.DurationMs = 5000 },
		"no slides":       func(r *LessonRecording) { r.Slides = nil },
		"slide id":        func(r *LessonRecording) { r.Slides[0].ID = "" },
	} {
		_, r := recordedLesson(t)
		mod(r)
		if err := r.CheckPaths(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
