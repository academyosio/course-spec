package coursespec

// CourseManifest is a course's course.json.
type CourseManifest struct {
	// SpecVersion selects the format version, such as "1".
	SpecVersion string `json:"specVersion" jsonschema:"pattern=^1(\\.[0-9]+)?$"`
	// ID is the course's stable kebab-case ID.
	ID string `json:"id" jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"`
	// Version is the course's semantic version.
	Version string `json:"version"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	// Description is a short Markdown description. Set it or DescriptionFile.
	Description string `json:"description,omitempty"`
	// DescriptionFile is a Markdown file holding a long description.
	DescriptionFile string `json:"descriptionFile,omitempty"`
	// Languages lists the languages the course is published in; the first is
	// the default. Empty means en-US only.
	Languages     []string `json:"languages,omitempty"`
	Audience      string   `json:"audience,omitempty"`
	Outcomes      []string `json:"outcomes,omitempty"`
	Prerequisites []string `json:"prerequisites,omitempty"`
	// Access is open, invite, or paid.
	Access    string   `json:"access,omitempty" jsonschema:"enum=open,enum=invite,enum=paid"`
	Tags      []string `json:"tags,omitempty"`
	Authors   []Author `json:"authors,omitempty"`
	Thumbnail string   `json:"thumbnail,omitempty"`
	// Theme is a directory of shared deck files, inside the same repository.
	Theme       string        `json:"theme,omitempty"`
	Certificate *Certificate  `json:"certificate,omitempty"`
	Video       *VideoPublish `json:"video,omitempty"`
	// Sections lists the course's sections in order.
	Sections []SectionRef `json:"sections"`
	Changes  []Change     `json:"changes,omitempty"`
}

// Author is a course author.
type Author struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Certificate rules.
const (
	RequiresAllQuizzes           = "allQuizzes"
	RequiresAllLessons           = "allLessons"
	RequiresAllLessonsAndQuizzes = "allLessonsAndQuizzes"
)

// Certificate describes the certificate a learner earns.
type Certificate struct {
	Enabled bool `json:"enabled"`
	// Requires is allQuizzes, allLessons, or allLessonsAndQuizzes.
	Requires string `json:"requires,omitempty" jsonschema:"enum=allQuizzes,enum=allLessons,enum=allLessonsAndQuizzes"`
	// Title defaults to the course title.
	Title string `json:"title,omitempty"`
	// Template defaults to the platform's template.
	Template string `json:"template,omitempty"`
}

// VideoPublish records how a course's videos are published.
type VideoPublish struct {
	// Provider is youtube or vimeo.
	Provider string `json:"provider" jsonschema:"enum=youtube,enum=vimeo"`
	// Visibility is public, unlisted, or private.
	Visibility string `json:"visibility,omitempty" jsonschema:"enum=public,enum=unlisted,enum=private"`
	// Playlist is each language's playlist title.
	Playlist map[string]string `json:"playlist,omitempty"`
	// DescriptionFooter is appended to each video's description, by language.
	DescriptionFooter map[string]string `json:"descriptionFooter,omitempty"`
}

// SectionRef is one entry in a course's section list.
type SectionRef struct {
	// Path is the section directory, relative to the course directory.
	Path string `json:"path"`
	// Released is false for a planned section that isn't published yet.
	// Absent means released.
	Released *bool `json:"released,omitempty"`
	// ID and Title name an unreleased section, whose directory may be
	// missing.
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

// IsReleased reports whether the section is published.
func (r SectionRef) IsReleased() bool { return r.Released == nil || *r.Released }

// Change is one entry in a course's change log.
type Change struct {
	Version string `json:"version"`
	Notes   string `json:"notes,omitempty"`
	// RetakeQuizzes names quizzes whose earlier passes are out of date.
	RetakeQuizzes []string `json:"retakeQuizzes,omitempty"`
}

// DefaultLanguage returns the course's default language.
func (m *CourseManifest) DefaultLanguage() string {
	if len(m.Languages) == 0 {
		return DefaultLanguage
	}
	return m.Languages[0]
}

// AllLanguages returns the course's languages, defaulting to en-US.
func (m *CourseManifest) AllLanguages() []string {
	if len(m.Languages) == 0 {
		return []string{DefaultLanguage}
	}
	return m.Languages
}
