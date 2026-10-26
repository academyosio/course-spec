package deck

import (
	"strings"
	"testing"
)

const sample = `<!doctype html><html><head><title>Sample Deck</title>
<link rel="stylesheet" href="../theme/deck.css?v=2">
<link rel="stylesheet" href="https://fonts.example.com/font.css">
<link rel="canonical" href="https://example.com/">
<script src="//cdn.example.com/lib.js"></script>
<style>@import "local.css"; body { background: url('img/bg.png'); } .x { background: url(data:image/png;base64,AA==) }</style>
</head><body>
<section class="slide" data-id="intro" data-title="Intro" data-vo="Hello   there. [PAUSE:500] Welcome. [PAUSE:250]"><h1>Intro</h1></section>
<section class="slide" data-id="first" data-lesson="First Lesson" data-vo="One.">
  <div class="slide" data-id="nested">nested slides are not counted</div>
  <img src="img/a.png"><a href="https://example.com/read-more">a link</a>
</section>
<section class="slide" data-id="second" data-lesson data-title="Second Title">
  <aside class="notes">From the
  notes.</aside>
</section>
</body></html>`

func mustParse(t *testing.T, html string) *Deck {
	t.Helper()
	d, err := Parse(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return d
}

func TestParse(t *testing.T) {
	d := mustParse(t, sample)
	if d.Title != "Sample Deck" {
		t.Errorf("title = %q", d.Title)
	}
	if got := strings.Join(d.SlideIDs(), ","); got != "intro,first,second" {
		t.Fatalf("slide IDs = %s", got)
	}
	intro := d.Slides[0]
	if intro.Narration != "Hello there. Welcome." || intro.PauseMs != 750 {
		t.Errorf("intro narration = %q pause %d", intro.Narration, intro.PauseMs)
	}
	if d.Slides[2].Narration != "From the notes." {
		t.Errorf("notes narration = %q", d.Slides[2].Narration)
	}
	if !d.Slides[1].StartsLesson || d.Slides[1].LessonTitle != "First Lesson" {
		t.Errorf("first slide lesson = %v %q", d.Slides[1].StartsLesson, d.Slides[1].LessonTitle)
	}
	if !strings.Contains(string(d.Slides[1].HTML), "nested slides") {
		t.Error("slide HTML misses its content")
	}
}

func TestParseNoSlides(t *testing.T) {
	if _, err := Parse(strings.NewReader("<html><body><p>none</p></body></html>")); err == nil {
		t.Fatal("expected an error for a page with no slides")
	}
}

func TestRefs(t *testing.T) {
	d := mustParse(t, sample)
	var remote, local []string
	for _, r := range d.Refs {
		switch {
		case r.Remote():
			remote = append(remote, r.URL)
		case r.Local():
			local = append(local, r.Path())
		}
	}
	if got := strings.Join(remote, " "); got != "https://fonts.example.com/font.css //cdn.example.com/lib.js" {
		t.Errorf("remote = %s", got)
	}
	if got := strings.Join(local, " "); got != "../theme/deck.css img/bg.png local.css img/a.png" {
		t.Errorf("local = %s", got)
	}
}

func TestLessons(t *testing.T) {
	d := mustParse(t, sample)
	ls := d.Lessons("section-id")
	if len(ls) != 2 {
		t.Fatalf("lessons = %d", len(ls))
	}
	if ls[0].ID != "first" || ls[0].Title != "First Lesson" || len(ls[0].Slides) != 2 {
		t.Errorf("lesson 1 = %s %q %d slides", ls[0].ID, ls[0].Title, len(ls[0].Slides))
	}
	if ls[1].ID != "second" || ls[1].Title != "Second Title" {
		t.Errorf("lesson 2 = %s %q", ls[1].ID, ls[1].Title)
	}
	if ls[0].Words() != 4 {
		t.Errorf("words = %d", ls[0].Words())
	}
}

func TestLessonsWithoutStarts(t *testing.T) {
	d := mustParse(t, `<title>Only</title><div class="slide" data-id="a"></div><div class="slide" data-id="b"></div>`)
	ls := d.Lessons("fallback")
	if len(ls) != 1 || ls[0].ID != "fallback" || ls[0].Title != "Only" || len(ls[0].Slides) != 2 {
		t.Fatalf("lessons = %+v", ls)
	}
}

func TestFingerprint(t *testing.T) {
	langs := []string{"es-ES", "en-US"}
	base := mustParse(t, sample).Lessons("s")[0]
	tr := Narration{"intro": {"es-ES": "Hola."}, "first": {"es-ES": "Uno."}}
	fp := base.Fingerprint(langs, "en-US", tr)
	if !strings.HasPrefix(fp, FingerprintPrefix) || len(fp) != len(FingerprintPrefix)+64 {
		t.Fatalf("fingerprint = %q", fp)
	}
	if got := base.Fingerprint([]string{"en-US", "es-ES"}, "en-US", tr); got != fp {
		t.Error("language order changed the fingerprint")
	}
	changedTr := Narration{"intro": {"es-ES": "Hola de nuevo."}, "first": {"es-ES": "Uno."}}
	if base.Fingerprint(langs, "en-US", changedTr) == fp {
		t.Error("a changed translation kept the fingerprint")
	}
	edited := mustParse(t, strings.Replace(sample, "<h1>Intro</h1>", "<h1>Introduction</h1>", 1)).Lessons("s")[0]
	if edited.Fingerprint(langs, "en-US", tr) == fp {
		t.Error("changed slide HTML kept the fingerprint")
	}
}

func TestTranscriptTextJoinsSegments(t *testing.T) {
	a := TranscriptLanguage{Segments: []TranscriptSegment{{Text: "One. Two."}}}
	b := TranscriptLanguage{Segments: []TranscriptSegment{{Text: "One."}, {Text: " Two."}}}
	if a.Text() != b.Text() {
		t.Errorf("%q != %q", a.Text(), b.Text())
	}
}

func TestMatch(t *testing.T) {
	d := mustParse(t, sample)
	byID := &Transcript{Slides: []TranscriptSlide{
		{ID: "first", Transcripts: map[string]TranscriptLanguage{"es-ES": {Segments: []TranscriptSegment{{Text: "Uno."}}}}},
		{ID: "gone", Transcripts: map[string]TranscriptLanguage{}},
	}}
	n, unknown, byPos := byID.Match(d)
	if byPos || n["first"]["es-ES"] != "Uno." || len(unknown) != 1 || unknown[0] != "gone" {
		t.Errorf("by ID: %v %v %v", n, unknown, byPos)
	}
	byIndex := &Transcript{Slides: []TranscriptSlide{
		{Index: 0, Transcripts: map[string]TranscriptLanguage{"es-ES": {Segments: []TranscriptSegment{{Text: "Hola."}}}}},
	}}
	n, _, byPos = byIndex.Match(d)
	if !byPos || n["intro"]["es-ES"] != "Hola." {
		t.Errorf("by position: %v %v", n, byPos)
	}
}

func TestFingerprintIgnoresSlideID(t *testing.T) {
	a := mustParse(t, sample).Lessons("s")[1]
	b := mustParse(t, strings.Replace(sample, `data-id="second"`, `data-id="renamed"`, 1)).Lessons("s")[1]
	if a.Fingerprint([]string{"en-US"}, "en-US", nil) != b.Fingerprint([]string{"en-US"}, "en-US", nil) {
		t.Error("renaming a slide changed the fingerprint")
	}
	if strings.Contains(string(a.Slides[0].HTML), "data-id") {
		t.Error("slide HTML kept its data-id")
	}
}
