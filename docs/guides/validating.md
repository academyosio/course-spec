# Validating a Course

## `coursespec validate`

```bash
coursespec validate <dir>                     # check a course; exits 1 on errors
coursespec validate --format json <dir>       # for tools and agents
coursespec validate --strict-narration <dir>  # also warn about characters TTS reads badly
```

Given a directory without a `course.json`, it checks every course beneath it,
skipping `.git`, `node_modules`, and `build` directories.

Each finding names its severity, the file relative to the course directory,
a JSON pointer or slide, the message, a code, and the fix:

```text
error: sections/01-designing-endpoints/section.json /lessons/2/id: lesson "pagination" is not in the deck [lesson-not-in-deck]
  fix: use the data-id of a slide with data-lesson, or set "type": "text" for a text lesson
testdata/courses/web-apis: 1 errors, 0 warnings
```

Errors make a course invalid and the command exit non-zero. Warnings point
at things that are probably wrong or will hurt learners, such as unnarrated
slides or an answer key a learner could pass by pattern.

The JSON output lists each course with its counts and findings:

```json
{
  "courses": [
    {
      "dir": "testdata/courses/web-apis",
      "errors": 1,
      "warnings": 0,
      "findings": [
        {
          "severity": "error",
          "code": "lesson-not-in-deck",
          "file": "sections/01-designing-endpoints/section.json",
          "pointer": "/lessons/2/id",
          "message": "lesson \"pagination\" is not in the deck",
          "fix": "use the data-id of a slide with data-lesson, or set \"type\": \"text\" for a text lesson"
        }
      ]
    }
  ]
}
```

Every code is listed in [Validation findings](../reference/findings.md).

## What it checks

1. Every manifest decodes, with no unknown properties.
2. IDs: course, section, lesson, and quiz IDs are lowercase kebab-case;
   lesson IDs are unique in the course, question IDs in their quiz.
3. Every listed section has a manifest, unless unreleased; every directory
   under `sections/` is listed; every section has a lesson or a quiz.
4. Each deck exists and parses, its slides have unique `data-id`s, every
   `lessons[]` ID that isn't a text lesson is in the deck, and each relative
   file the deck loads exists.
5. Quiz rules from [Quizzes](quizzes.md).
6. Every ID in `published.json` still exists or is retired.
7. Content changed without a version bump.
8. Referenced files exist, and the theme stays inside the repository.
9. Transcripts: every slide `id` is in the deck, and every slide has
   narration in each of the course's languages.

## `coursespec lessons`

```bash
coursespec lessons <dir>
coursespec lessons --format json <dir>
```

Lists each section, lesson, and quiz with its status from
[published state](published-state.md): `new`, `changed`, `published`,
`retired`, or `planned`. It needs a course that loads, so run `validate`
first when it reports a problem.

## `coursespec schema`

```bash
coursespec schema           # list the schemas: course, published, quiz, section
coursespec schema quiz      # print one
```

## In CI

```yaml
- run: go install github.com/academyosio/course-spec/cmd/coursespec@v0.1.0
- run: coursespec validate courses/
```
