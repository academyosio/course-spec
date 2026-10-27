package coursespec

// Publish status of a lesson or quiz, compared with published.json.
const (
	StatusNew       = "new"       // never published
	StatusChanged   = "changed"   // content differs from what was published
	StatusPublished = "published" // unchanged since published
	StatusRetired   = "retired"   // kept for its published ID
	StatusPlanned   = "planned"   // in an unreleased section
)

// Report lists a course's sections, lessons, and quizzes with their status.
type Report struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Version  string          `json:"version"`
	Dir      string          `json:"dir"`
	Sections []SectionReport `json:"sections"`
}

// SectionReport is one section in a Report.
type SectionReport struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Released bool           `json:"released"`
	Lessons  []LessonReport `json:"lessons,omitempty"`
	Quiz     *QuizReport    `json:"quiz,omitempty"`
}

// LessonReport is one lesson in a Report.
type LessonReport struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Slides      int    `json:"slides,omitempty"`
	Status      string `json:"status"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// QuizReport is a section's quiz in a Report.
type QuizReport struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Questions   int    `json:"questions"`
	Status      string `json:"status"`
	Fingerprint string `json:"fingerprint"`
}

// Status reports each lesson and quiz as new, changed, published, retired,
// or planned, by comparing content fingerprints with published.json. It
// works after a fresh clone, because published.json is committed.
func (c *Course) Status() Report {
	r := Report{ID: c.Manifest.ID, Title: c.Manifest.Title, Version: c.Manifest.Version, Dir: c.Dir}
	pub := c.Published
	if pub == nil {
		pub = &Published{}
	}
	for _, s := range c.Sections {
		sr := SectionReport{ID: s.ID(), Title: s.Title(), Released: s.Ref.IsReleased()}
		for _, l := range s.Lessons {
			lr := LessonReport{ID: l.ID, Title: l.Title, Type: l.Type, Slides: len(l.Slides), Fingerprint: c.Fingerprint(s, l)}
			p, published := pub.Lessons[l.ID]
			switch {
			case l.Retired:
				lr.Status = StatusRetired
			case !sr.Released:
				lr.Status = StatusPlanned
			case !published:
				lr.Status = StatusNew
			case p.ContentFingerprint != "" && lr.Fingerprint != "" && p.ContentFingerprint != lr.Fingerprint:
				lr.Status = StatusChanged
			default:
				lr.Status = StatusPublished
			}
			sr.Lessons = append(sr.Lessons, lr)
		}
		if s.Quiz != nil {
			qr := &QuizReport{ID: s.Quiz.ID, Title: s.Quiz.Title, Questions: len(s.Quiz.Questions), Fingerprint: QuizFingerprint(s.Quiz)}
			p, published := pub.Quizzes[s.Quiz.ID]
			switch {
			case !sr.Released:
				qr.Status = StatusPlanned
			case !published:
				qr.Status = StatusNew
			case p.Fingerprint != "" && p.Fingerprint != qr.Fingerprint:
				qr.Status = StatusChanged
			default:
				qr.Status = StatusPublished
			}
			sr.Quiz = qr
		}
		r.Sections = append(r.Sections, sr)
	}
	return r
}
