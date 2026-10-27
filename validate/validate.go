package validate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/deck"
)

var (
	idRe     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	semverRe = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
)

const (
	wordsPerMinute   = 150
	longLessonMinute = 10
)

type checker struct {
	dir      string
	opts     Options
	findings []Finding
	manifest *coursespec.CourseManifest
	// sections that loaded, for cross-section checks
	sections []*section
	lessons  map[string]string // lesson ID -> file that defines it
	quizzes  map[string]string // quiz ID -> file
}

type section struct {
	ref          coursespec.SectionRef
	dir          string
	manifest     *coursespec.SectionManifest
	deck         *deck.Deck
	translations deck.Narration
	lessons      []coursespec.Lesson
	quiz         *coursespec.Quiz
	quizFile     string
}

// Course checks the course directory dir and returns every finding.
func Course(dir string, opts Options) []Finding {
	c := &checker{dir: dir, opts: opts, lessons: map[string]string{}, quizzes: map[string]string{}}
	c.run()
	return c.findings
}

func (c *checker) add(sev, code, file, ptr, msg, fix string) {
	c.findings = append(c.findings, Finding{Severity: sev, Code: code, File: file, Pointer: ptr, Message: msg, Fix: fix})
}

func (c *checker) rel(path string) string {
	r, err := filepath.Rel(c.dir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(r)
}

func (c *checker) decodeFinding(file string, err error) {
	var de *coursespec.DecodeError
	ptr := ""
	msg := err.Error()
	if errors.As(err, &de) {
		ptr = de.Pointer
		msg = de.Err.Error()
	}
	fix := "correct the JSON"
	if strings.Contains(msg, "unknown field") {
		fix = "remove the property or correct its spelling; property names are camelCase"
	}
	c.add(Error, "invalid-json", c.rel(file), ptr, msg, fix)
}

func (c *checker) run() {
	path := filepath.Join(c.dir, coursespec.CourseFile)
	m, err := coursespec.ReadCourseManifest(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.add(Error, "missing-course", coursespec.CourseFile, "", "the directory has no course.json", "create course.json, or run in a directory that contains one")
			return
		}
		c.decodeFinding(path, err)
		return
	}
	c.manifest = m
	c.checkCourse()
	c.checkSectionDirs()
	for i, ref := range m.Sections {
		c.checkSection(i, ref)
	}
	for _, s := range c.sections {
		if s.quiz != nil {
			c.checkQuiz(s.quizFile, s.quiz)
		}
	}
	c.checkChanges()
	c.checkPublished()
}

func (c *checker) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c *checker) checkCourse() {
	m := c.manifest
	const f = coursespec.CourseFile
	if !idRe.MatchString(m.ID) {
		c.add(Error, "invalid-id", f, "/id", fmt.Sprintf("course ID %q is not lowercase kebab-case", m.ID), "use lowercase letters, digits, and single hyphens, such as \"building-web-apis\"")
	}
	if !semverRe.MatchString(m.Version) {
		c.add(Error, "invalid-version", f, "/version", fmt.Sprintf("version %q is not a semantic version", m.Version), "use MAJOR.MINOR.PATCH, such as \"1.0.0\"")
	}
	if strings.TrimSpace(m.Title) == "" {
		c.add(Error, "missing-title", f, "/title", "the course has no title", "set \"title\"")
	}
	if m.Description != "" && m.DescriptionFile != "" {
		c.add(Error, "description-twice", f, "/descriptionFile", "both description and descriptionFile are set", "keep one: description for a short text, descriptionFile for a long one")
	}
	c.checkFile(f, "/descriptionFile", m.DescriptionFile)
	c.checkFile(f, "/thumbnail", m.Thumbnail)
	if m.Access != "" && !slices.Contains([]string{"open", "invite", "paid"}, m.Access) {
		c.add(Error, "invalid-enum", f, "/access", fmt.Sprintf("access %q is not valid", m.Access), "use open, invite, or paid")
	}
	seen := map[string]bool{}
	for i, l := range m.Languages {
		if seen[l] {
			c.add(Error, "duplicate-language", f, fmt.Sprintf("/languages/%d", i), fmt.Sprintf("language %q is listed twice", l), "list each language once")
		}
		seen[l] = true
	}
	if cert := m.Certificate; cert != nil {
		if cert.Requires != "" && !slices.Contains([]string{coursespec.RequiresAllQuizzes, coursespec.RequiresAllLessons, coursespec.RequiresAllLessonsAndQuizzes}, cert.Requires) {
			c.add(Error, "invalid-enum", f, "/certificate/requires", fmt.Sprintf("requires %q is not valid", cert.Requires), "use allQuizzes, allLessons, or allLessonsAndQuizzes")
		}
		c.checkFile(f, "/certificate/template", cert.Template)
	}
	if v := m.Video; v != nil {
		if !slices.Contains([]string{"youtube", "vimeo"}, v.Provider) {
			c.add(Error, "invalid-enum", f, "/video/provider", fmt.Sprintf("provider %q is not valid", v.Provider), "use youtube or vimeo")
		}
		if v.Visibility != "" && !slices.Contains([]string{"public", "unlisted", "private"}, v.Visibility) {
			c.add(Error, "invalid-enum", f, "/video/visibility", fmt.Sprintf("visibility %q is not valid", v.Visibility), "use public, unlisted, or private")
		}
		langs := m.AllLanguages()
		for _, field := range []struct {
			name string
			m    map[string]string
		}{{"playlist", v.Playlist}, {"descriptionFooter", v.DescriptionFooter}} {
			for _, l := range sortedKeys(field.m) {
				if !slices.Contains(langs, l) {
					c.add(Warning, "unknown-language", f, "/video/"+field.name+"/"+l, fmt.Sprintf("language %q is not in the course's languages", l), "add it to \"languages\" or remove the entry")
				}
			}
		}
	}
	c.checkTheme()
	ppath := filepath.Join(c.dir, coursespec.PronunciationsFile)
	if c.exists(ppath) {
		if _, err := coursespec.ReadPronunciations(ppath); err != nil {
			c.decodeFinding(ppath, err)
		}
	}
	if len(m.Sections) == 0 {
		c.add(Error, "no-sections", f, "/sections", "the course lists no sections", "add each section directory to \"sections\" as {\"path\": \"sections/01-...\"}")
	}
}

func (c *checker) checkFile(file, ptr, rel string) {
	if rel == "" {
		return
	}
	if !c.inside(filepath.Join(c.dir, rel), c.dir) {
		c.add(Error, "path-outside-course", file, ptr, fmt.Sprintf("%q is outside the course directory", rel), "keep course files inside the course directory")
		return
	}
	if !c.exists(filepath.Join(c.dir, rel)) {
		c.add(Error, "missing-file", file, ptr, fmt.Sprintf("%q does not exist", rel), "create the file or correct the path; paths are relative to the course directory")
	}
}

func (c *checker) inside(path, root string) bool {
	r, err := filepath.Rel(root, path)
	return err == nil && r != ".." && !strings.HasPrefix(r, ".."+string(filepath.Separator))
}

func (c *checker) checkTheme() {
	theme := c.manifest.Theme
	if theme == "" {
		return
	}
	abs, err := filepath.Abs(filepath.Join(c.dir, theme))
	if err != nil {
		c.add(Error, "unreadable", coursespec.CourseFile, "/theme", err.Error(), "")
		return
	}
	if root := repoRoot(c.dir); root != "" && !c.inside(abs, root) {
		c.add(Error, "theme-outside-repository", coursespec.CourseFile, "/theme", fmt.Sprintf("theme %q is outside the repository", theme), "keep the theme in the same repository as the course")
		return
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		c.add(Error, "missing-theme", coursespec.CourseFile, "/theme", fmt.Sprintf("theme directory %q does not exist", theme), "create the directory or correct the path, relative to the course directory")
	}
}

// repoRoot returns the nearest directory at or above dir holding .git.
func repoRoot(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil {
			return abs
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}

func (c *checker) checkSectionDirs() {
	listed := map[string]bool{}
	for _, ref := range c.manifest.Sections {
		listed[filepath.Clean(ref.Path)] = true
	}
	entries, err := os.ReadDir(filepath.Join(c.dir, "sections"))
	if err != nil {
		return // no sections directory: section paths are checked one by one
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join("sections", e.Name())
		if !listed[p] {
			c.add(Error, "unlisted-section", coursespec.CourseFile, "/sections", fmt.Sprintf("directory %q is not listed in sections", filepath.ToSlash(p)), fmt.Sprintf("add {\"path\": %q} to \"sections\", or remove the directory", filepath.ToSlash(p)))
		}
	}
}

func (c *checker) defineLesson(id, file, ptr string) {
	if prev, ok := c.lessons[id]; ok {
		c.add(Error, "duplicate-lesson-id", file, ptr, fmt.Sprintf("lesson ID %q is also used in %s", id, prev), "lesson IDs must be unique within the course; rename one")
		return
	}
	c.lessons[id] = file
}

func (c *checker) checkSection(i int, ref coursespec.SectionRef) {
	ptr := fmt.Sprintf("/sections/%d", i)
	const cf = coursespec.CourseFile
	if ref.Path == "" {
		c.add(Error, "missing-path", cf, ptr+"/path", "the section has no path", "set \"path\" to the section directory, such as \"sections/01-intro\"")
		return
	}
	sdir := filepath.Join(c.dir, ref.Path)
	if !c.inside(sdir, c.dir) {
		c.add(Error, "path-outside-course", cf, ptr+"/path", fmt.Sprintf("%q is outside the course directory", ref.Path), "keep sections inside the course directory")
		return
	}
	if !ref.IsReleased() {
		if !idRe.MatchString(ref.ID) {
			c.add(Error, "invalid-id", cf, ptr+"/id", fmt.Sprintf("unreleased section ID %q is not lowercase kebab-case", ref.ID), "give an unreleased section an \"id\" in kebab-case")
		}
		if strings.TrimSpace(ref.Title) == "" {
			c.add(Error, "missing-title", cf, ptr+"/title", "the unreleased section has no title", "give an unreleased section a \"title\"")
		}
		c.sections = append(c.sections, &section{ref: ref, dir: sdir})
		return
	}
	if ref.ID != "" || ref.Title != "" {
		c.add(Warning, "unused-property", cf, ptr, "id and title in sections are read only for unreleased sections", "set them in the section's section.json")
	}
	mpath := filepath.Join(sdir, coursespec.SectionFile)
	sm, err := coursespec.ReadSectionManifest(mpath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.add(Error, "missing-section", cf, ptr+"/path", fmt.Sprintf("%s does not exist", c.rel(mpath)), "create the section's section.json, or mark the section \"released\": false")
			return
		}
		c.decodeFinding(mpath, err)
		return
	}
	sf := c.rel(mpath)
	s := &section{ref: ref, dir: sdir, manifest: sm}
	for _, other := range c.sections {
		if other.id() == sm.ID {
			c.add(Error, "duplicate-section-id", sf, "/id", fmt.Sprintf("section ID %q is used twice", sm.ID), "give each section its own ID")
		}
	}
	c.sections = append(c.sections, s)
	if !idRe.MatchString(sm.ID) {
		c.add(Error, "invalid-id", sf, "/id", fmt.Sprintf("section ID %q is not lowercase kebab-case", sm.ID), "use lowercase letters, digits, and single hyphens")
	}
	if strings.TrimSpace(sm.Title) == "" {
		c.add(Error, "missing-title", sf, "/title", "the section has no title", "set \"title\"")
	}

	switch sm.Source.Type {
	case coursespec.SourceHTMLDeck:
		c.checkDeck(s, sf)
	case coursespec.SourceNone:
		if sm.Source.Path != "" {
			c.add(Warning, "unused-property", sf, "/source/path", "a none source has no file", "remove \"path\"")
		}
	case coursespec.SourceMarp, coursespec.SourceVideo:
		c.add(Error, "unsupported-source", sf, "/source/type", fmt.Sprintf("source type %q is not supported yet", sm.Source.Type), "use htmlDeck, or none for text lessons")
		return
	default:
		c.add(Error, "invalid-enum", sf, "/source/type", fmt.Sprintf("source type %q is not valid", sm.Source.Type), "use htmlDeck or none")
		return
	}

	lessons, problems := coursespec.ResolveLessons(sm, s.deck)
	for _, p := range problems {
		c.add(Error, "lesson-not-in-deck", sf, fmt.Sprintf("/lessons/%d/id", p.Index), p.Message, p.Fix)
	}
	s.lessons = lessons
	c.checkLessonEntries(s, sf)

	qpath, err := coursespec.QuizPath(sdir, sm)
	hasQuiz := err == nil && qpath != ""
	if err != nil {
		c.add(Error, "unreadable", sf, "/quiz", err.Error(), "")
	} else if qpath != "" {
		q, err := coursespec.ReadQuiz(qpath)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			c.add(Error, "missing-file", sf, "/quiz", fmt.Sprintf("quiz file %q does not exist", sm.Quiz), "create it or correct the path, relative to the section directory")
		case err != nil:
			c.decodeFinding(qpath, err)
		default:
			s.quiz, s.quizFile = q, c.rel(qpath)
			if prev, ok := c.quizzes[q.ID]; ok {
				c.add(Error, "duplicate-quiz-id", s.quizFile, "/id", fmt.Sprintf("quiz ID %q is also used in %s", q.ID, prev), "give each quiz its own ID")
			} else {
				c.quizzes[q.ID] = s.quizFile
			}
		}
	}
	if len(s.lessons) == 0 && !hasQuiz {
		c.add(Error, "empty-section", sf, "", "the section has no lessons and no quiz", "add lessons, or a quiz for a quiz-only section")
	}
}

func (s *section) id() string {
	if s.manifest != nil {
		return s.manifest.ID
	}
	return s.ref.ID
}

func (c *checker) checkLessonEntries(s *section, sf string) {
	sm := s.manifest
	for i, e := range sm.Lessons {
		ptr := fmt.Sprintf("/lessons/%d", i)
		if !idRe.MatchString(e.ID) {
			c.add(Error, "invalid-id", sf, ptr+"/id", fmt.Sprintf("lesson ID %q is not lowercase kebab-case", e.ID), "use lowercase letters, digits, and single hyphens")
		}
		if e.Type != "" && e.Type != coursespec.LessonVideo && e.Type != coursespec.LessonText {
			c.add(Error, "invalid-enum", sf, ptr+"/type", fmt.Sprintf("lesson type %q is not valid", e.Type), "use video or text")
		}
		text := e.Type == coursespec.LessonText || (sm.Source.Type == coursespec.SourceNone && e.Type == "")
		if sm.Source.Type == coursespec.SourceNone && e.Type == coursespec.LessonVideo {
			c.add(Error, "video-without-source", sf, ptr+"/type", "a section with a none source has only text lessons", "set \"type\": \"text\", or give the section a deck")
		}
		if text {
			if e.Content == "" {
				c.add(Error, "missing-content", sf, ptr+"/content", fmt.Sprintf("text lesson %q has no content file", e.ID), "set \"content\" to a Markdown file, such as \"text/glossary.md\"")
			} else if !c.exists(filepath.Join(s.dir, e.Content)) {
				c.add(Error, "missing-file", sf, ptr+"/content", fmt.Sprintf("%q does not exist", e.Content), "create the file or correct the path, relative to the section directory")
			}
			if strings.TrimSpace(e.Title) == "" {
				c.add(Error, "missing-title", sf, ptr+"/title", fmt.Sprintf("text lesson %q has no title", e.ID), "set \"title\"")
			}
		}
		if e.After != "" {
			if !text {
				c.add(Warning, "unused-property", sf, ptr+"/after", "after places text lessons only", "remove \"after\"")
			} else if !slices.ContainsFunc(s.lessons, func(l coursespec.Lesson) bool { return l.ID == e.After && l.ID != e.ID }) {
				c.add(Error, "unknown-lesson", sf, ptr+"/after", fmt.Sprintf("after names %q, which is not a lesson in this section", e.After), "name a lesson in the same section")
			}
		}
		if e.File != "" {
			c.add(Warning, "unused-property", sf, ptr+"/file", "file is read only for video sources", "remove \"file\"")
		}
	}
	for _, l := range s.lessons {
		c.defineLesson(l.ID, sf, "")
	}
}

func (c *checker) checkDeck(s *section, sf string) {
	sm := s.manifest
	if sm.Source.Path == "" {
		c.add(Error, "missing-path", sf, "/source/path", "the htmlDeck source has no path", "set \"path\" to the deck, such as \"deck/index.html\"")
		return
	}
	dpath := filepath.Join(s.dir, sm.Source.Path)
	df := c.rel(dpath)
	d, err := deck.ParseFile(dpath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.add(Error, "missing-file", sf, "/source/path", fmt.Sprintf("deck %q does not exist", sm.Source.Path), "create the deck or correct the path, relative to the section directory")
			return
		}
		c.add(Error, "invalid-deck", df, "", err.Error(), "give each slide class=\"slide\"")
		return
	}
	s.deck = d

	seen := map[string]int{}
	var unnamed []int
	for _, sl := range d.Slides {
		if sl.ID == "" {
			unnamed = append(unnamed, sl.Index+1)
			continue
		}
		if first, ok := seen[sl.ID]; ok {
			c.add(Error, "duplicate-slide-id", df, "#"+sl.ID, fmt.Sprintf("slides %d and %d both have data-id %q", first, sl.Index+1, sl.ID), "give each slide its own data-id")
		} else {
			seen[sl.ID] = sl.Index + 1
		}
		if sl.StartsLesson && !idRe.MatchString(sl.ID) {
			c.add(Error, "invalid-id", df, "#"+sl.ID, fmt.Sprintf("lesson ID %q (a data-id with data-lesson) is not lowercase kebab-case", sl.ID), "use lowercase letters, digits, and single hyphens")
		}
	}
	if len(unnamed) > 0 {
		c.add(Error, "missing-slide-id", df, "", fmt.Sprintf("slides %s have no data-id", joinInts(unnamed)), "give every slide a stable data-id; lessons and translations are matched by it")
	}

	var remote, missing []string
	for _, r := range d.Refs {
		switch {
		case r.Remote():
			remote = append(remote, r.URL)
		case r.Local():
			if !c.exists(filepath.Join(filepath.Dir(dpath), filepath.FromSlash(r.Path()))) {
				missing = append(missing, r.Path())
			}
		}
	}
	for _, p := range uniq(missing) {
		c.add(Error, "missing-deck-file", df, "", fmt.Sprintf("the deck loads %q, which does not exist", p), "add the file or correct the path, relative to the deck")
	}
	for _, u := range uniq(remote) {
		c.add(Warning, "remote-deck-file", df, "", fmt.Sprintf("the deck loads %q from another host", u), "download the file into the course or theme and load it by a relative path; sandboxed and offline players block remote files")
	}

	var silent []string
	for _, sl := range d.Slides {
		if sl.Narration == "" {
			silent = append(silent, slideName(sl))
		}
		if c.opts.StrictNarration && strings.ContainsAny(sl.Narration, "()/→") {
			c.add(Warning, "narration-characters", df, "#"+sl.ID, "the narration contains characters TTS reads badly: ( ) / or →", "write them out in words")
		}
	}
	if len(silent) > 0 {
		c.add(Warning, "no-narration", df, "", fmt.Sprintf("slides without narration: %s", strings.Join(silent, ", ")), "add data-vo to each slide, or aside class=\"notes\"")
	}
	for _, l := range d.Lessons(sm.ID) {
		if mins := l.Words() / wordsPerMinute; mins > longLessonMinute {
			c.add(Warning, "long-lesson", df, "#"+l.ID, fmt.Sprintf("lesson %q runs about %d minutes", l.ID, mins), "split it into shorter lessons with data-lesson")
		}
	}
	c.checkTranscript(s, d, filepath.Join(filepath.Dir(dpath), coursespec.TranscriptFile))
}

func (c *checker) checkTranscript(s *section, d *deck.Deck, tpath string) {
	langs := c.manifest.AllLanguages()
	def := c.manifest.DefaultLanguage()
	if !c.exists(tpath) {
		if len(langs) > 1 {
			c.add(Warning, "missing-translation", c.rel(filepath.Dir(tpath)), "", fmt.Sprintf("there is no transcript, so the deck has no narration in %s", strings.Join(langs[1:], ", ")), "add voiceover/transcript.json with each slide's narration in every language")
		}
		return
	}
	tf := c.rel(tpath)
	t, err := deck.ReadTranscript(tpath)
	if err != nil {
		c.add(Error, "invalid-json", tf, "", err.Error(), "correct the JSON")
		return
	}
	n, unknown, byPos := t.Match(d)
	s.translations = n
	if byPos {
		c.add(Warning, "transcript-by-position", tf, "", "slide entries have no id, so translations are matched by position", "add \"id\" to each slide entry, equal to the slide's data-id")
	}
	for _, id := range unknown {
		c.add(Error, "unknown-slide", tf, "", fmt.Sprintf("slide %q is not in the deck", id), "rename the entry to the slide's current data-id, or remove it")
	}
	var differ []string
	missing := map[string][]string{}
	for _, sl := range d.Slides {
		if sl.ID == "" {
			continue
		}
		if txt, ok := n[sl.ID][def]; ok && txt != sl.Narration {
			differ = append(differ, sl.ID)
		}
		for _, l := range langs[1:] {
			if strings.TrimSpace(n[sl.ID][l]) == "" {
				missing[l] = append(missing[l], sl.ID)
			}
		}
	}
	if len(differ) > 0 {
		c.add(Warning, "transcript-differs", tf, "", fmt.Sprintf("the %s narration differs from the deck on slides: %s", def, strings.Join(differ, ", ")), "the deck is the source for the default language; regenerate the transcript from the deck")
	}
	for _, l := range sortedKeys(missing) {
		c.add(Warning, "missing-translation", tf, "", fmt.Sprintf("slides without %s narration: %s", l, strings.Join(missing[l], ", ")), "translate them; recording tools record a language only when every slide has it")
	}
}

func (c *checker) checkChanges() {
	for i, ch := range c.manifest.Changes {
		if !semverRe.MatchString(ch.Version) {
			c.add(Error, "invalid-version", coursespec.CourseFile, fmt.Sprintf("/changes/%d/version", i), fmt.Sprintf("version %q is not a semantic version", ch.Version), "use MAJOR.MINOR.PATCH")
		}
		for j, q := range ch.RetakeQuizzes {
			if _, ok := c.quizzes[q]; !ok {
				c.add(Error, "unknown-quiz", coursespec.CourseFile, fmt.Sprintf("/changes/%d/retakeQuizzes/%d", i, j), fmt.Sprintf("quiz %q does not exist", q), "name a quiz ID from a quiz file")
			}
		}
	}
}

func slideName(s deck.Slide) string {
	if s.ID != "" {
		return s.ID
	}
	return fmt.Sprintf("slide %d", s.Index+1)
}

func joinInts(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, ", ")
}

func uniq(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
