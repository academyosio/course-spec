package validate_test

import (
	"strings"
	"testing"

	"github.com/academyosio/course-spec/internal/testcourse"
	"github.com/academyosio/course-spec/validate"
)

const (
	sec1    = "sections/01-designing-endpoints/"
	deckRel = sec1 + "deck/index.html"
	quiz1   = sec1 + "quiz.json"
	trRel   = sec1 + "deck/voiceover/transcript.json"
)

func TestValidCourse(t *testing.T) {
	fs := validate.Course(testcourse.Source(), validate.Options{})
	if len(fs) != 0 {
		t.Fatalf("findings for the valid course:\n%s", list(fs))
	}
}

func TestMissingCourse(t *testing.T) {
	fs := validate.Course(t.TempDir(), validate.Options{})
	expect(t, fs, validate.Error, "missing-course")
}

type tc struct {
	name     string
	severity string
	code     string
	edit     func(t *testing.T, dir string)
	opts     validate.Options
}

func TestRules(t *testing.T) {
	cases := []tc{
		{"unknown property", validate.Error, "invalid-json", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["titel"] = "x" })
		}, validate.Options{}},
		{"unsupported spec version", validate.Error, "invalid-json", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["specVersion"] = "2" })
		}, validate.Options{}},
		{"course ID", validate.Error, "invalid-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["id"] = "Web_APIs" })
		}, validate.Options{}},
		{"version", validate.Error, "invalid-version", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["version"] = "1.0" })
		}, validate.Options{}},
		{"access enum", validate.Error, "invalid-enum", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["access"] = "public" })
		}, validate.Options{}},
		{"description twice", validate.Error, "description-twice", func(t *testing.T, d string) {
			testcourse.Write(t, d, "description.md", "Long.")
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) {
				m["description"] = "Short."
				m["descriptionFile"] = "description.md"
			})
		}, validate.Options{}},
		{"missing thumbnail", validate.Error, "missing-file", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) { m["thumbnail"] = "assets/none.png" })
		}, validate.Options{}},
		{"theme outside course still in repository is fine, missing theme is not", validate.Error, "missing-theme", func(t *testing.T, d string) {
			testcourse.Remove(t, d, "theme")
		}, validate.Options{}},
		{"unlisted section", validate.Error, "unlisted-section", func(t *testing.T, d string) {
			testcourse.Write(t, d, "sections/09-forgotten/section.json", `{}`)
		}, validate.Options{}},
		{"missing section manifest", validate.Error, "missing-section", func(t *testing.T, d string) {
			testcourse.Remove(t, d, "sections/02-final-assessment/section.json")
		}, validate.Options{}},
		{"unreleased section ID", validate.Error, "invalid-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) {
				m["sections"].([]any)[2].(map[string]any)["id"] = ""
			})
		}, validate.Options{}},
		{"duplicate section ID", validate.Error, "duplicate-section-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "sections/02-final-assessment/section.json", func(m map[string]any) { m["id"] = "designing-endpoints" })
		}, validate.Options{}},
		{"empty section", validate.Error, "empty-section", func(t *testing.T, d string) {
			testcourse.Remove(t, d, "sections/02-final-assessment/quiz.json")
		}, validate.Options{}},
		{"unsupported source", validate.Error, "unsupported-source", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "sections/02-final-assessment/section.json", func(m map[string]any) {
				m["source"] = map[string]any{"type": "marp", "path": "deck.md"}
			})
		}, validate.Options{}},
		{"lesson not in deck", validate.Error, "lesson-not-in-deck", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, sec1+"section.json", func(m map[string]any) {
				m["lessons"] = append(m["lessons"].([]any), map[string]any{"id": "renamed"})
			})
		}, validate.Options{}},
		{"text lesson without content", validate.Error, "missing-file", func(t *testing.T, d string) {
			testcourse.Remove(t, d, sec1+"text/glossary.md")
		}, validate.Options{}},
		{"after names no lesson", validate.Error, "unknown-lesson", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, sec1+"section.json", func(m map[string]any) {
				m["lessons"].([]any)[1].(map[string]any)["after"] = "nowhere"
			})
		}, validate.Options{}},
		{"duplicate lesson ID", validate.Error, "duplicate-lesson-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, sec1+"section.json", func(m map[string]any) {
				m["lessons"].([]any)[1].(map[string]any)["id"] = "methods"
			})
		}, validate.Options{}},
		{"missing deck", validate.Error, "missing-file", func(t *testing.T, d string) {
			testcourse.Remove(t, d, deckRel)
		}, validate.Options{}},
		{"duplicate slide ID", validate.Error, "duplicate-slide-id", func(t *testing.T, d string) {
			replace(t, d, deckRel, `data-id="methods"`, `data-id="resources"`)
		}, validate.Options{}},
		{"missing slide ID", validate.Error, "missing-slide-id", func(t *testing.T, d string) {
			replace(t, d, deckRel, `data-id="title" `, ``)
		}, validate.Options{}},
		{"lesson ID not kebab-case", validate.Error, "invalid-id", func(t *testing.T, d string) {
			replace(t, d, deckRel, `data-id="methods"`, `data-id="HTTP_Methods"`)
		}, validate.Options{}},
		{"missing deck file", validate.Error, "missing-deck-file", func(t *testing.T, d string) {
			testcourse.Remove(t, d, "theme/deck.css")
		}, validate.Options{}},
		{"remote deck file", validate.Warning, "remote-deck-file", func(t *testing.T, d string) {
			replace(t, d, deckRel, `</head>`, `<script src="https://cdn.example.com/x.js"></script></head>`)
		}, validate.Options{}},
		{"slide without narration", validate.Warning, "no-narration", func(t *testing.T, d string) {
			replace(t, d, deckRel, `data-vo="An API is a set of resources."`, ``)
		}, validate.Options{}},
		{"strict narration", validate.Warning, "narration-characters", func(t *testing.T, d string) {
			replace(t, d, deckRel, `An API is a set of resources.`, `An API (or two) is a set of resources.`)
		}, validate.Options{StrictNarration: true}},
		{"long lesson", validate.Warning, "long-lesson", func(t *testing.T, d string) {
			replace(t, d, deckRel, `An API is a set of resources.`, strings.Repeat("word ", 2000))
		}, validate.Options{}},
		{"transcript slide not in deck", validate.Error, "unknown-slide", func(t *testing.T, d string) {
			replace(t, d, trRel, `"id": "methods"`, `"id": "verbs"`)
		}, validate.Options{}},
		{"transcript by position", validate.Warning, "transcript-by-position", func(t *testing.T, d string) {
			for _, id := range []string{"title", "resources", "methods"} {
				replace(t, d, trRel, `"id": "`+id+`", `, ``)
			}
		}, validate.Options{}},
		{"missing translation", validate.Warning, "missing-translation", func(t *testing.T, d string) {
			replace(t, d, trRel, `{"text": "Una API es un conjunto de recursos."}`, `{"text": ""}`)
		}, validate.Options{}},
		{"default language differs", validate.Warning, "transcript-differs", func(t *testing.T, d string) {
			replace(t, d, trRel, `Welcome to designing endpoints.`, `Welcome, everyone.`)
		}, validate.Options{}},
		{"no correct option", validate.Error, "no-correct-option", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) {
				for _, o := range testcourse.Question(m, 0)["options"].([]any) {
					o.(map[string]any)["correct"] = false
				}
			})
		}, validate.Options{}},
		{"true/false without answer", validate.Error, "missing-answer", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { delete(testcourse.Question(m, 1), "correctAnswer") })
		}, validate.Options{}},
		{"short answer without answers", validate.Error, "missing-answer", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 2)["acceptedAnswers"] = []any{} })
		}, validate.Options{}},
		{"wrong answer field", validate.Error, "wrong-answer-field", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 0)["correctAnswer"] = true })
		}, validate.Options{}},
		{"missing explanation", validate.Error, "missing-explanation", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 1)["explanation"] = "" })
		}, validate.Options{}},
		{"question names unknown lesson", validate.Error, "unknown-lesson", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 0)["lesson"] = "nope" })
		}, validate.Options{}},
		{"duplicate question ID", validate.Error, "duplicate-question-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 1)["id"] = "q1-idempotent" })
		}, validate.Options{}},
		{"pass percentage", validate.Error, "out-of-range", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { m["passPercentage"] = 120 })
		}, validate.Options{}},
		{"question type", validate.Error, "invalid-enum", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { testcourse.Question(m, 1)["type"] = "true_false" })
		}, validate.Options{}},
		{"answer pattern", validate.Warning, "answer-pattern", func(t *testing.T, d string) {
			q := func(id string) map[string]any {
				return map[string]any{"id": id, "type": "multipleChoice", "prompt": "Pick", "explanation": "Because.",
					"options": []any{map[string]any{"id": "a", "text": "A", "correct": true}, map[string]any{"id": "b", "text": "B", "correct": false}}}
			}
			testcourse.EditJSON(t, d, quiz1, func(m map[string]any) { m["questions"] = []any{q("q1"), q("q2"), q("q3")} })
		}, validate.Options{}},
		{"duplicate quiz ID", validate.Error, "duplicate-quiz-id", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "sections/02-final-assessment/quiz.json", func(m map[string]any) { m["id"] = "designing-endpoints-quiz" })
		}, validate.Options{}},
		{"retake names unknown quiz", validate.Error, "unknown-quiz", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) {
				m["changes"] = []any{map[string]any{"version": "1.0.0", "retakeQuizzes": []any{"nope"}}}
			})
		}, validate.Options{}},
		{"published lesson removed", validate.Error, "published-id-removed", func(t *testing.T, d string) {
			testcourse.Write(t, d, "published.json", `{"lessons": {"gone": {}}}`)
		}, validate.Options{}},
		{"published quiz removed", validate.Error, "published-id-removed", func(t *testing.T, d string) {
			testcourse.Write(t, d, "published.json", `{"quizzes": {"gone-quiz": {}}}`)
		}, validate.Options{}},
		{"unversioned change", validate.Warning, "unversioned-change", func(t *testing.T, d string) {
			testcourse.Write(t, d, "published.json", `{"courseVersion": "1.0.0", "lessons": {"methods": {"contentFingerprint": "sha256:old"}}}`)
		}, validate.Options{}},
		{"playlist language", validate.Warning, "unknown-language", func(t *testing.T, d string) {
			testcourse.EditJSON(t, d, "course.json", func(m map[string]any) {
				m["video"] = map[string]any{"provider": "youtube", "playlist": map[string]any{"fr-FR": "API"}}
			})
		}, validate.Options{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := testcourse.Copy(t)
			c.edit(t, dir)
			expect(t, validate.Course(dir, c.opts), c.severity, c.code)
		})
	}
}

func TestRenameDetected(t *testing.T) {
	dir := testcourse.Copy(t)
	// Publish "methods" with its current fingerprint, then rename its slide.
	before := validate.Course(dir, validate.Options{})
	if len(before) != 0 {
		t.Fatal(list(before))
	}
	fp := fingerprintOf(t, dir, "methods")
	testcourse.Write(t, dir, "published.json", `{"lessons": {"methods": {"contentFingerprint": "`+fp+`"}}}`)
	replace(t, dir, deckRel, `data-id="methods"`, `data-id="http-methods"`)
	replace(t, dir, trRel, `"id": "methods"`, `"id": "http-methods"`)
	replace(t, dir, quiz1, `"lesson": "methods"`, `"lesson": "http-methods"`)
	fs := validate.Course(dir, validate.Options{})
	f := expect(t, fs, validate.Error, "published-id-removed")
	if !strings.Contains(f.Message, `renamed to "http-methods"`) {
		t.Errorf("message = %s", f.Message)
	}
}

func TestCounts(t *testing.T) {
	e, w := validate.Counts([]validate.Finding{{Severity: validate.Error}, {Severity: validate.Warning}, {Severity: validate.Warning}})
	if e != 1 || w != 2 {
		t.Errorf("counts = %d, %d", e, w)
	}
}

func expect(t *testing.T, fs []validate.Finding, severity, code string) validate.Finding {
	t.Helper()
	for _, f := range fs {
		if f.Code == code && f.Severity == severity {
			if f.Message == "" || f.File == "" {
				t.Errorf("finding lacks a message or file: %+v", f)
			}
			return f
		}
	}
	t.Fatalf("no %s %q; findings:\n%s", severity, code, list(fs))
	return validate.Finding{}
}

func list(fs []validate.Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.String() + "\n")
	}
	return b.String()
}

func replace(t *testing.T, dir, rel, old, new string) {
	t.Helper()
	s := testcourse.Read(t, dir, rel)
	if !strings.Contains(s, old) {
		t.Fatalf("%s does not contain %q", rel, old)
	}
	testcourse.Write(t, dir, rel, strings.Replace(s, old, new, 1))
}
