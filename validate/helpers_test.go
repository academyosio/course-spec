package validate_test

import (
	"testing"

	coursespec "github.com/academyosio/course-spec"
)

func fingerprintOf(t *testing.T, dir, lessonID string) string {
	t.Helper()
	c, err := coursespec.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range c.Status().Sections {
		for _, l := range s.Lessons {
			if l.ID == lessonID {
				return l.Fingerprint
			}
		}
	}
	t.Fatalf("no lesson %q", lessonID)
	return ""
}
