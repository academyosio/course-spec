package validate

import (
	"fmt"
	"path/filepath"
	"strings"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/deck"
)

func (c *checker) checkPublished() {
	path := filepath.Join(c.dir, coursespec.PublishedFile)
	if !c.exists(path) {
		return
	}
	pub, err := coursespec.ReadPublished(path)
	if err != nil {
		c.decodeFinding(path, err)
		return
	}
	const f = coursespec.PublishedFile
	current := map[string]string{} // lesson ID -> fingerprint
	for _, s := range c.sections {
		for _, l := range s.lessons {
			current[l.ID] = c.fingerprint(s, l)
		}
	}
	for _, id := range sortedKeys(pub.Lessons) {
		if _, ok := current[id]; ok {
			continue
		}
		msg := fmt.Sprintf("published lesson %q no longer exists", id)
		if fp := pub.Lessons[id].ContentFingerprint; fp != "" {
			for newID, cur := range current {
				if _, published := pub.Lessons[newID]; !published && cur == fp {
					msg += fmt.Sprintf("; it looks renamed to %q", newID)
				}
			}
		}
		c.add(Error, "published-id-removed", f, "/lessons/"+id, msg, "restore the ID, or keep the lesson in section.json with \"retired\": true; published IDs key learner progress")
	}
	for _, id := range sortedKeys(pub.Quizzes) {
		if _, ok := c.quizzes[id]; !ok {
			c.add(Error, "published-id-removed", f, "/quizzes/"+id, fmt.Sprintf("published quiz %q no longer exists", id), "restore the quiz ID")
		}
	}
	for _, name := range sortedKeys(pub.Platforms) {
		p := pub.Platforms[name]
		for _, id := range sortedKeys(p.Lessons) {
			if _, ok := current[id]; !ok {
				c.add(Error, "published-id-removed", f, "/platforms/"+name+"/lessons/"+id, fmt.Sprintf("lesson %q, published to %s, no longer exists", id, name), "restore the ID, or retire the lesson")
			}
		}
		for _, id := range sortedKeys(p.Quizzes) {
			if _, ok := c.quizzes[id]; !ok {
				c.add(Error, "published-id-removed", f, "/platforms/"+name+"/quizzes/"+id, fmt.Sprintf("quiz %q, published to %s, no longer exists", id, name), "restore the quiz ID")
			}
		}
	}
	if pub.CourseVersion == "" || pub.CourseVersion != c.manifest.Version {
		return
	}
	var changed []string
	for _, id := range sortedKeys(pub.Lessons) {
		fp, cur := pub.Lessons[id].ContentFingerprint, current[id]
		if fp != "" && cur != "" && fp != cur {
			changed = append(changed, id)
		}
	}
	for _, s := range c.sections {
		if s.quiz == nil {
			continue
		}
		if pq, ok := pub.Quizzes[s.quiz.ID]; ok && pq.Fingerprint != "" && pq.Fingerprint != coursespec.QuizFingerprint(s.quiz) {
			changed = append(changed, s.quiz.ID)
		}
	}
	if len(changed) > 0 {
		c.add(Warning, "unversioned-change", coursespec.CourseFile, "/version", fmt.Sprintf("content changed since version %s was published: %s", pub.CourseVersion, strings.Join(changed, ", ")), "raise \"version\": minor for new or rewritten lessons, patch for corrections")
	}
}

func (c *checker) fingerprint(s *section, l coursespec.Lesson) string {
	if len(l.Slides) == 0 {
		return ""
	}
	dl := deck.Lesson{ID: l.ID, Title: l.Title, Slides: l.Slides}
	return dl.Fingerprint(c.manifest.AllLanguages(), c.manifest.DefaultLanguage(), s.translations)
}
