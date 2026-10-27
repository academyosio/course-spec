# Quizzes

A section's quiz, with its answer key, is part of the course, in one format
every platform reads. It sits in the section directory as `quiz.json`, or the
file the section's `quiz` names.

```json
{
  "id": "designing-endpoints-quiz",
  "title": "Designing Endpoints Quiz",
  "passPercentage": 70,
  "allowRetry": true,
  "maxAttempts": 0,
  "showFeedback": true,
  "shuffleOptions": true,
  "questions": [
    {
      "id": "q1-idempotent-methods",
      "type": "multipleChoice",
      "lesson": "methods",
      "prompt": "Which methods are idempotent?",
      "options": [
        {"id": "a", "text": "GET", "correct": true},
        {"id": "b", "text": "POST", "correct": false},
        {"id": "c", "text": "PUT", "correct": true}
      ],
      "explanation": "GET and PUT can be repeated with the same result; POST can't.",
      "points": 1
    },
    {
      "id": "q2-404",
      "type": "trueFalse",
      "prompt": "A 404 means the server failed.",
      "correctAnswer": false,
      "explanation": "404 means the resource wasn't found, a client-side condition."
    },
    {
      "id": "q3-status-code",
      "type": "shortAnswer",
      "prompt": "Which status code means \"Created\"?",
      "acceptedAnswers": ["201", "201 Created"],
      "explanation": "201 Created, with a Location header for the new resource."
    }
  ]
}
```

## Settings

| Property | Meaning |
| --- | --- |
| `id`, `title` | Required. The ID is kebab-case and unique in the course |
| `passPercentage` | 0 to 100 |
| `allowRetry` | Whether a learner may try again |
| `maxAttempts` | 0 means unlimited |
| `showFeedback` | Whether explanations are shown after grading |
| `shuffleOptions` | Asks the platform to shuffle multiple-choice options for each learner; without it, options appear in file order |
| `questions` | At least one |

## Questions

Every question has an `id` unique in the quiz, a `type`, a `prompt`, and an
`explanation`. The explanation is required: feedback without a reason
doesn't teach. `points` defaults to 1. `lesson`, when set, names the lesson
the question covers, in any section, so a final assessment can point each
question back at what to review.

| `type` | Answer key | Rules |
| --- | --- | --- |
| `multipleChoice` | `options`, each with `id`, `text`, and `correct` | At least two options with unique IDs and at least one correct. With more than one correct, the learner must choose exactly the correct set |
| `trueFalse` | `correctAnswer`, `true` or `false` | Required |
| `shortAnswer` | `acceptedAnswers` | One or more, compared ignoring case and repeated spaces |

A question carrying another type's answer field, such as `correctAnswer` on a
multiple-choice question, is an error, because it is almost always a
mistake.

## Patterns a learner could pass by

Without `shuffleOptions`, the validator warns when three or more
single-answer multiple-choice questions all have their correct option in the
same position. Vary where the correct option sits, or set `shuffleOptions`.

## Final assessments

Put a final assessment in a quiz-only section (a `none` source with no
lessons), as [The course directory](course-directory.md#quiz-only-sections)
shows. Its questions can name lessons in any section.

## Changes

When a quiz changes enough that earlier passes shouldn't count, list it in
`course.json`'s `changes[].retakeQuizzes` for the new version. Each quiz also
has a fingerprint in `published.json`, so tools notice a changed quiz the way
they notice a changed lesson.
