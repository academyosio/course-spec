package deck

// Lesson is a run of consecutive slides, recorded as one video.
type Lesson struct {
	// ID is the starting slide's data-id. For a deck with no lesson starts
	// it is the fallback ID given to Lessons, usually the section's ID.
	ID string
	// Title is the data-lesson value, else the starting slide's data-title,
	// else its ID.
	Title string
	// Slides are the lesson's slides in order.
	Slides []Slide
}

// Lessons divides the deck into lessons. A lesson starts at each slide with
// data-lesson; slides before the first start join the first lesson. A deck
// with no lesson starts is one lesson with fallbackID and the deck's title.
func (d *Deck) Lessons(fallbackID string) []Lesson {
	var lessons []Lesson
	started := false // whether the current lesson has its start slide
	for _, s := range d.Slides {
		switch {
		case len(lessons) == 0:
			lessons = append(lessons, Lesson{})
		case s.StartsLesson && started:
			lessons = append(lessons, Lesson{})
			started = false
		}
		l := &lessons[len(lessons)-1]
		if s.StartsLesson && !started {
			l.ID = s.ID
			l.Title = firstNonEmpty(s.LessonTitle, s.Title, s.ID)
			started = true
		}
		l.Slides = append(l.Slides, s)
	}
	if len(lessons) == 1 && !started {
		lessons[0].ID = fallbackID
		lessons[0].Title = firstNonEmpty(d.Title, fallbackID)
	}
	return lessons
}

// Words counts the words in the lesson's default-language narration.
func (l Lesson) Words() int {
	n := 0
	for _, s := range l.Slides {
		n += countWords(s.Narration)
	}
	return n
}

func countWords(s string) int {
	n, in := 0, false
	for _, r := range s {
		space := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if !space && !in {
			n++
		}
		in = !space
	}
	return n
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
