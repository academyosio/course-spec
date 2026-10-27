package validate

import (
	"fmt"
	"strings"

	coursespec "github.com/academyosio/course-spec"
)

func (c *checker) checkQuiz(f string, q *coursespec.Quiz) {
	if !idRe.MatchString(q.ID) {
		c.add(Error, "invalid-id", f, "/id", fmt.Sprintf("quiz ID %q is not lowercase kebab-case", q.ID), "use lowercase letters, digits, and single hyphens")
	}
	if strings.TrimSpace(q.Title) == "" {
		c.add(Error, "missing-title", f, "/title", "the quiz has no title", "set \"title\"")
	}
	if q.PassPercentage < 0 || q.PassPercentage > 100 {
		c.add(Error, "out-of-range", f, "/passPercentage", fmt.Sprintf("passPercentage %d is not between 0 and 100", q.PassPercentage), "use a percentage from 0 to 100")
	}
	if q.MaxAttempts < 0 {
		c.add(Error, "out-of-range", f, "/maxAttempts", "maxAttempts is negative", "use 0 for unlimited attempts")
	}
	if len(q.Questions) == 0 {
		c.add(Error, "no-questions", f, "/questions", "the quiz has no questions", "add questions")
	}
	ids := map[string]bool{}
	positions := map[int]int{} // correct option position -> count, single-answer questions
	single := 0
	for i, qu := range q.Questions {
		p := fmt.Sprintf("/questions/%d", i)
		if qu.ID == "" {
			c.add(Error, "missing-id", f, p+"/id", "the question has no id", "give each question an id unique in the quiz")
		} else if ids[qu.ID] {
			c.add(Error, "duplicate-question-id", f, p+"/id", fmt.Sprintf("question ID %q is used twice", qu.ID), "give each question its own id")
		}
		ids[qu.ID] = true
		if strings.TrimSpace(qu.Prompt) == "" {
			c.add(Error, "missing-prompt", f, p+"/prompt", "the question has no prompt", "set \"prompt\"")
		}
		if strings.TrimSpace(qu.Explanation) == "" {
			c.add(Error, "missing-explanation", f, p+"/explanation", "the question has no explanation", "say why the answer is right; feedback without a reason doesn't teach")
		}
		if qu.Points != nil && *qu.Points < 0 {
			c.add(Error, "out-of-range", f, p+"/points", "points is negative", "use 0 or more; it defaults to 1")
		}
		if qu.Lesson != "" {
			if _, ok := c.lessons[qu.Lesson]; !ok {
				c.add(Error, "unknown-lesson", f, p+"/lesson", fmt.Sprintf("lesson %q does not exist", qu.Lesson), "name a lesson ID from any section, or remove \"lesson\"")
			}
		}
		switch qu.Type {
		case coursespec.QuestionMultipleChoice:
			c.checkOptions(f, p, qu)
			if qu.CorrectAnswer != nil {
				c.add(Error, "wrong-answer-field", f, p+"/correctAnswer", "correctAnswer is for trueFalse questions", "mark correct options with \"correct\": true")
			}
			if len(qu.AcceptedAnswers) > 0 {
				c.add(Error, "wrong-answer-field", f, p+"/acceptedAnswers", "acceptedAnswers is for shortAnswer questions", "remove it")
			}
			correct := -1
			n := 0
			for j, o := range qu.Options {
				if o.Correct {
					correct = j
					n++
				}
			}
			if n == 1 {
				single++
				positions[correct]++
			}
		case coursespec.QuestionTrueFalse:
			if qu.CorrectAnswer == nil {
				c.add(Error, "missing-answer", f, p+"/correctAnswer", "the trueFalse question has no correctAnswer", "set \"correctAnswer\" to true or false")
			}
			if len(qu.Options) > 0 || len(qu.AcceptedAnswers) > 0 {
				c.add(Error, "wrong-answer-field", f, p, "a trueFalse question uses only correctAnswer", "remove options and acceptedAnswers")
			}
		case coursespec.QuestionShortAnswer:
			ok := false
			for _, a := range qu.AcceptedAnswers {
				if strings.TrimSpace(a) != "" {
					ok = true
				}
			}
			if !ok {
				c.add(Error, "missing-answer", f, p+"/acceptedAnswers", "the shortAnswer question has no accepted answers", "list one or more in \"acceptedAnswers\"")
			}
			if len(qu.Options) > 0 || qu.CorrectAnswer != nil {
				c.add(Error, "wrong-answer-field", f, p, "a shortAnswer question uses only acceptedAnswers", "remove options and correctAnswer")
			}
		default:
			c.add(Error, "invalid-enum", f, p+"/type", fmt.Sprintf("question type %q is not valid", qu.Type), "use multipleChoice, trueFalse, or shortAnswer")
		}
	}
	if !q.ShuffleOptions && single >= 3 && len(positions) == 1 {
		for pos := range positions {
			c.add(Warning, "answer-pattern", f, "/questions", fmt.Sprintf("the correct option is always number %d", pos+1), "vary where the correct option sits, or set \"shuffleOptions\": true")
		}
	}
}

func (c *checker) checkOptions(f, p string, qu coursespec.Question) {
	if len(qu.Options) < 2 {
		c.add(Error, "too-few-options", f, p+"/options", "a multipleChoice question needs at least two options", "add options")
	}
	ids := map[string]bool{}
	correct := 0
	for j, o := range qu.Options {
		op := fmt.Sprintf("%s/options/%d", p, j)
		if o.ID == "" {
			c.add(Error, "missing-id", f, op+"/id", "the option has no id", "give each option an id, such as \"a\"")
		} else if ids[o.ID] {
			c.add(Error, "duplicate-option-id", f, op+"/id", fmt.Sprintf("option ID %q is used twice", o.ID), "give each option its own id")
		}
		ids[o.ID] = true
		if strings.TrimSpace(o.Text) == "" {
			c.add(Error, "missing-text", f, op+"/text", "the option has no text", "set \"text\"")
		}
		if o.Correct {
			correct++
		}
	}
	if len(qu.Options) >= 2 && correct == 0 {
		c.add(Error, "no-correct-option", f, p+"/options", "no option is marked correct", "set \"correct\": true on the right option or options")
	}
}
