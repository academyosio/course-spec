# Review Suggestions: Final Assessments, Practice Quizzes, Partial Release

From the PBHQ Academy course authors, after planning certificates for
*From Launch to Traction*: six sections, each ending with a quiz, then a final
assessment. Three gaps in the spec, in order of need.

## 1. A quiz-only section for a final assessment (needed now)

The TRD's Quiz section already expects "a review quiz at the end of a course",
whose questions point at lessons in any section. But a quiz belongs to a
section, and every section needs a source with lessons, so there is no
declared place for a course-level final assessment.

**Suggestion:** allow a section with `source.type` `none`, no lessons, and a
quiz. State it in the Section Manifest rules, and make the validator accept it.

```json
{
  "id": "final-assessment",
  "title": "Final Assessment",
  "summary": "Scenario questions across the whole course.",
  "source": {"type": "none"},
  "quiz": "quiz.json"
}
```

Rules worth adding:

- A section must have at least one lesson or a quiz.
- In a quiz-only section, each question may set `lesson` to a lesson in any
  section, as the Quiz rules already allow.
- A platform shows a quiz-only section as one quiz item after the previous
  section, with no video.

On AcademyOS, a quiz is attached to a lesson, so the import would create a
quiz lesson in that section, as `course.json` does today with `"type": "quiz"`
lessons.

## 2. Practice quizzes (later)

Some courses want section quizzes as practice, with only the final assessment
counting toward the certificate. With unlimited retries and feedback on, a
section quiz is a learning aid. Requiring it for the certificate adds little
once a final assessment tests the whole course.

**Suggestion:** one optional quiz field, `practice` (default `false`). A
practice quiz is shown and graded with feedback, but `certificate.requires:
allQuizzes` (and `allLessonsAndQuizzes`) counts only quizzes that are not
practice. No new `requires` values are needed.

```json
{"id": "measure-before-you-market-quiz", "practice": true, "passPercentage": 70, "questions": []}
```

Validator: warn when `requires` is `allQuizzes` and every quiz is practice,
because then the certificate needs no quiz at all.

AcademyOS today: `certificateEligible` requires a passed result for every quiz
in the course. It would skip quizzes marked practice, and the quiz model would
need the flag.

## 3. Releasing sections before the course is complete (needed at launch)

A course is often published a section at a time. *From Launch to Traction*
will launch with its first sections, and later sections follow as they are
written. Some depend on weeks of real data. The manifest lists sections by
path, and a missing directory is an error, so a course can't declare its full
outline while only part of it exists.

**Suggestion:** let a course manifest list planned sections that aren't
released yet:

```json
"sections": [
  "sections/01-from-launch-to-traction",
  {"path": "sections/02-measure-before-you-market", "released": false}
]
```

- An unreleased section needs only an ID and title, and its directory may be
  missing. Validation skips it apart from ID uniqueness.
- A platform shows it as "coming soon", without lessons or quiz.
- While any section is unreleased, a platform issues no certificates, even
  when `certificate.enabled` is true, and says on the course page that the
  certificate becomes available when the course is complete.

Without this, authors either hide the outline until the course is finished
or turn certificates on by hand at the right moment. The second is easy to
get wrong in both directions.
