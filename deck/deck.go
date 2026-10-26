// Package deck reads HTML slide decks in the videoascode conventions: each
// slide is an element with class "slide" and a data-id, narration is in
// data-vo or a child aside.notes, and a slide with data-lesson starts a
// lesson.
//
// It depends only on an HTML parser, so recording tools and platforms can
// import it for lesson discovery without browser automation.
package deck

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Slide is one slide of a deck.
type Slide struct {
	// Index is the slide's zero-based position in the deck.
	Index int
	// ID is the slide's data-id, or its id attribute. It may be empty.
	ID string
	// Title is the slide's data-title. It may be empty.
	Title string
	// Narration is the default-language narration, normalized: [PAUSE:ms]
	// directives removed and whitespace collapsed.
	Narration string
	// PauseMs is the total of the narration's [PAUSE:ms] directives.
	PauseMs int
	// StartsLesson reports whether the slide has data-lesson.
	StartsLesson bool
	// LessonTitle is the data-lesson value.
	LessonTitle string
	// HTML is the slide element rendered back to HTML without its own
	// data-id, so renaming a slide doesn't change its content.
	HTML []byte
}

// Deck is a parsed HTML deck.
type Deck struct {
	// Title is the page's <title>.
	Title  string
	Slides []Slide
	// Refs are the files the page loads: scripts, stylesheets, images,
	// media, and CSS url() references.
	Refs []Ref
}

// Ref is a file a deck loads.
type Ref struct {
	// URL is the reference as written.
	URL string
	// Element names where it was found, such as "script src" or "style url()".
	Element string
}

// Remote reports whether the reference loads from another host.
func (r Ref) Remote() bool {
	u := strings.ToLower(strings.TrimSpace(r.URL))
	return strings.HasPrefix(u, "http:") || strings.HasPrefix(u, "https:") || strings.HasPrefix(u, "//")
}

// Local reports whether the reference is a relative file path, as opposed
// to a remote URL, a data: URL, or a fragment.
func (r Ref) Local() bool {
	u := strings.TrimSpace(r.URL)
	if u == "" || r.Remote() || strings.HasPrefix(u, "#") {
		return false
	}
	if i := strings.IndexAny(u, ":/?#"); i > 0 && u[i] == ':' {
		return false // a scheme such as data: or blob:
	}
	return !strings.HasPrefix(u, "/")
}

// Path returns a local reference's file path, without query or fragment.
func (r Ref) Path() string {
	u := strings.TrimSpace(r.URL)
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	return u
}

var (
	pauseRe      = regexp.MustCompile(`\[PAUSE:(\d+)\]`)
	whitespaceRe = regexp.MustCompile(`\s+`)
	cssURLRe     = regexp.MustCompile(`url\(\s*['"]?([^'")\s]+)['"]?\s*\)`)
	cssImportRe  = regexp.MustCompile(`@import\s+['"]([^'"]+)['"]`)
)

// ParseFile reads the deck at path.
func ParseFile(path string) (*Deck, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d, err := Parse(f)
	if cerr := f.Close(); err == nil && cerr != nil {
		err = cerr
	}
	return d, err
}

// Parse reads an HTML deck. Slides nested in another slide aren't counted.
// It returns an error if the page has no slides.
func Parse(r io.Reader) (*Deck, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	d := &Deck{}
	var walk func(n *html.Node, inSlide bool) error
	walk = func(n *html.Node, inSlide bool) error {
		if n.Type == html.ElementNode {
			if n.Data == "title" && d.Title == "" {
				d.Title = Normalize(textContent(n))
			}
			d.collectRefs(n)
			if !inSlide && hasClass(n, "slide") {
				s, err := readSlide(n, len(d.Slides))
				if err != nil {
					return err
				}
				d.Slides = append(d.Slides, s)
				inSlide = true
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c, inSlide); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(doc, false); err != nil {
		return nil, err
	}
	if len(d.Slides) == 0 {
		return nil, fmt.Errorf(`no elements with class "slide"`)
	}
	return d, nil
}

func (d *Deck) collectRefs(n *html.Node) {
	add := func(url, element string) {
		if strings.TrimSpace(url) != "" {
			d.Refs = append(d.Refs, Ref{URL: url, Element: element})
		}
	}
	switch n.Data {
	case "script", "img", "iframe", "video", "audio", "source", "track", "embed":
		add(attr(n, "src"), n.Data+" src")
		if n.Data == "video" {
			add(attr(n, "poster"), "video poster")
		}
	case "link":
		rel := strings.Fields(strings.ToLower(attr(n, "rel")))
		if slices.Contains(rel, "stylesheet") || slices.Contains(rel, "icon") || slices.Contains(rel, "preload") || slices.Contains(rel, "modulepreload") {
			add(attr(n, "href"), "link href")
		}
	case "object":
		add(attr(n, "data"), "object data")
	case "style":
		css := textContent(n)
		for _, m := range cssURLRe.FindAllStringSubmatch(css, -1) {
			add(m[1], "style url()")
		}
		for _, m := range cssImportRe.FindAllStringSubmatch(css, -1) {
			add(m[1], "style @import")
		}
	}
	if style := attr(n, "style"); style != "" {
		for _, m := range cssURLRe.FindAllStringSubmatch(style, -1) {
			add(m[1], n.Data+" style url()")
		}
	}
}

func readSlide(n *html.Node, index int) (Slide, error) {
	s := Slide{Index: index, Title: Normalize(attr(n, "data-title")), ID: attr(n, "data-id")}
	if s.ID == "" {
		s.ID = attr(n, "id")
	}
	raw, ok := attrOK(n, "data-vo")
	if !ok {
		if notes := findNotes(n); notes != nil {
			raw = textContent(notes)
		}
	}
	for _, m := range pauseRe.FindAllStringSubmatch(raw, -1) {
		ms, err := strconv.Atoi(m[1])
		if err != nil {
			return Slide{}, fmt.Errorf("slide %d: pause %q: %w", index+1, m[0], err)
		}
		s.PauseMs += ms
	}
	s.Narration = Normalize(raw)
	s.LessonTitle, s.StartsLesson = attrOK(n, "data-lesson")
	s.LessonTitle = Normalize(s.LessonTitle)
	var buf bytes.Buffer
	if err := html.Render(&buf, withoutAttr(n, "data-id")); err != nil {
		return Slide{}, fmt.Errorf("slide %d: render: %w", index+1, err)
	}
	s.HTML = buf.Bytes()
	return s, nil
}

// withoutAttr returns a shallow copy of n without the attribute key. The
// copy shares n's children, so it is only for rendering.
func withoutAttr(n *html.Node, key string) *html.Node {
	c := *n
	c.Parent, c.PrevSibling, c.NextSibling = nil, nil, nil
	c.Attr = nil
	for _, a := range n.Attr {
		if a.Namespace != "" || a.Key != key {
			c.Attr = append(c.Attr, a)
		}
	}
	return &c
}

// Normalize removes [PAUSE:ms] directives and collapses whitespace, the form
// narration is compared and hashed in.
func Normalize(s string) string {
	return strings.TrimSpace(whitespaceRe.ReplaceAllString(pauseRe.ReplaceAllString(s, " "), " "))
}

// SlideIDs returns each slide's ID in order; slides without one are empty.
func (d *Deck) SlideIDs() []string {
	ids := make([]string, len(d.Slides))
	for i, s := range d.Slides {
		ids[i] = s.ID
	}
	return ids
}

// Slide returns the slide with the given ID.
func (d *Deck) Slide(id string) (Slide, bool) {
	for _, s := range d.Slides {
		if s.ID == id && id != "" {
			return s, true
		}
	}
	return Slide{}, false
}

func hasClass(n *html.Node, class string) bool {
	return slices.Contains(strings.Fields(attr(n, "class")), class)
}

func findNotes(n *html.Node) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "aside" && hasClass(c, "notes") {
			return c
		}
		if found := findNotes(c); found != nil {
			return found
		}
	}
	return nil
}

func attr(n *html.Node, key string) string {
	v, ok := attrOK(n, key)
	if !ok {
		return ""
	}
	return v
}

func attrOK(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Namespace == "" && a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}
