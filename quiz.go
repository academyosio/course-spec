package coursespec

// Question types.
const (
	QuestionMultipleChoice = "multipleChoice"
	QuestionTrueFalse      = "trueFalse"
	QuestionShortAnswer    = "shortAnswer"
)

// Quiz is a section's quiz with its answer key.
type Quiz struct {
	ID             string `json:"id" jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"`
	Title          string `json:"title"`
	PassPercentage int    `json:"passPercentage" jsonschema:"minimum=0,maximum=100"`
	AllowRetry     bool   `json:"allowRetry,omitempty"`
	// MaxAttempts of 0 means unlimited.
	MaxAttempts  int  `json:"maxAttempts,omitempty" jsonschema:"minimum=0"`
	ShowFeedback bool `json:"showFeedback,omitempty"`
	// ShuffleOptions asks the platform to shuffle multiple-choice options.
	ShuffleOptions bool       `json:"shuffleOptions,omitempty"`
	Questions      []Question `json:"questions"`
}

// Question is one quiz question. Which answer fields apply depends on Type.
type Question struct {
	ID string `json:"id"`
	// Type is multipleChoice, trueFalse, or shortAnswer.
	Type string `json:"type" jsonschema:"enum=multipleChoice,enum=trueFalse,enum=shortAnswer"`
	// Lesson names the lesson the question covers, in any section.
	Lesson string `json:"lesson,omitempty"`
	Prompt string `json:"prompt"`
	// Options are a multiple-choice question's choices.
	Options []Option `json:"options,omitempty"`
	// CorrectAnswer is a true/false question's answer.
	CorrectAnswer *bool `json:"correctAnswer,omitempty"`
	// AcceptedAnswers are a short-answer question's answers, compared
	// ignoring case and repeated spaces.
	AcceptedAnswers []string `json:"acceptedAnswers,omitempty"`
	// Explanation says why the answer is right. It is required.
	Explanation string `json:"explanation"`
	// Points defaults to 1.
	Points *int `json:"points,omitempty" jsonschema:"minimum=0"`
}

// Option is one choice in a multiple-choice question.
type Option struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

// PointsOrDefault returns the question's points, 1 when unset.
func (q Question) PointsOrDefault() int {
	if q.Points == nil {
		return 1
	}
	return *q.Points
}
