# course-spec: Technical Requirements

## Module

- Module path: `github.com/academyosio/course-spec`
- Go package: `coursespec` at the module root, holding the types and the loader.
- Subpackages:

  | Package | Purpose |
  | --- | --- |
  | `coursespec` | Types for every file, and `Load(dir)` which resolves a course directory into one `Course` value |
  | `validate` | Schema and cross-reference checks, returning findings with file, path, message, and fix |
  | `deck` | Discovers slides and lessons in an HTML deck (`data-id`, `data-lesson`, `data-vo`) |
  | `schema` | Generated JSON Schema files, embedded with `//go:embed` |
  | `cmd/coursespec` | Cobra CLI: `validate`, `lessons`, `schema` |

- Dependencies are kept small so platforms can import the module: `invopop/jsonschema` (generation only), a YAML decoder, and `golang.org/x/net/html` for deck parsing.

## Directory Layout

```
<course>/
  course.yaml               course manifest (or course.json)
  published.json            published state, written by tools, committed
  terms.json                pronunciations for narration (optional)
  theme/                    shared deck CSS, JS, fonts (optional)
  assets/                   thumbnail and course images (optional)
  sections/
    01-<slug>/
      section.yaml          section manifest (or section.json)
      deck/index.html       media source: slides with narration
      quiz.yaml             quiz with answer key (optional; or quiz.json)
      lessons/*.md          text lessons (optional)
      build/                generated: clips, videos, captions, lessons.json (not committed)
```

Rules:

- Section directories are listed in the course manifest in order. A directory under `sections/` that the manifest doesn't list is an error, so a forgotten section can't go unnoticed.
- The `NN-` prefix is for people browsing the tree; order comes from the manifest. IDs come from the manifests, not from directory names.
- A file may be YAML (`.yaml`) or JSON (`.json`), not both for the same manifest.
- `build/` holds generated files only and belongs in `.gitignore`.

## Course Manifest

```yaml
specVersion: "1"
id: building-web-apis
version: 1.2.0
title: Building Web APIs
summary: Design, build, and operate an HTTP API.
description: |
  Longer description in Markdown.
language: en-US
audience: Backend developers who have built a web app.
outcomes:
  - Design resource-oriented endpoints
  - Return errors clients can act on
prerequisites: []            # course IDs
access: open                 # open | invite | paid
tags: [apis, http]
authors:
  - name: Jane Author
    url: https://example.com
thumbnail: assets/thumbnail.png
certificate:
  enabled: true
  requires: allQuizzes       # allQuizzes | allLessons | allLessonsAndQuizzes
sections:
  - sections/01-designing-endpoints
  - sections/02-handling-errors
changes:
  - version: 1.2.0
    notes: Rewrote the pagination lesson.
    retakeQuizzes: [designing-endpoints-quiz]
```

- `specVersion` is required and selects the schema. A loader refuses a major version it doesn't know.
- `version` is semantic: major for restructuring, minor for new or rewritten lessons, patch for corrections.
- `changes[].retakeQuizzes` names quizzes whose earlier passes a platform should treat as out of date.

## Section Manifest

```yaml
id: designing-endpoints
title: Designing Endpoints
summary: Resources, methods, and status codes.
source:
  type: htmlDeck             # htmlDeck | marp | video | none
  path: deck/index.html
lessons:                     # details for discovered lessons, by ID
  - id: resources
    objectives:
      - Model a domain as resources
    free: true
quiz: quiz.yaml              # default: quiz.yaml or quiz.json if present
```

Sources:

| `source.type` | Lessons come from | Video comes from |
| --- | --- | --- |
| `htmlDeck` | Slides with `data-lesson`; the lesson ID is the starting slide's `data-id` | Recorded from the deck |
| `marp` | Slides marked as lesson starts (convention defined when supported) | Recorded from the deck |
| `video` | `lessons[]`, each with `file` (relative path, not committed) or an already-published video | The file |
| `none` | `lessons[]`, text lessons only | None |

Lesson fields:

| Field | Meaning |
| --- | --- |
| `id` | Required. Unique within the course |
| `title` | Required when not discovered; overrides a discovered title |
| `type` | `video` (default with a media source) or `text` |
| `content` | Path to a Markdown file, for text lessons |
| `objectives` | Learning objectives |
| `free` | Free preview |
| `estimatedMinutes` | Used until a recording gives the real duration |
| `retired` | Kept to preserve a published ID; not shown to new learners |

An entry in `lessons[]` whose ID the deck doesn't contain is an error, so a renamed slide is caught. Text lessons may be placed among discovered ones with `after: <lesson id>`.

## HTML Deck Conventions

The spec adopts the videoascode conventions so the same deck records and validates without changes:

- A slide is an element with class `slide` and a `data-id`, unique in the deck.
- A slide with `data-lesson` starts a lesson. The value is the title; when empty, the slide's `data-title` or ID is used.
- Slides before the first lesson start belong to the first lesson. A deck with no lesson starts is one lesson, with the section's ID.
- Narration is in `data-vo`, or in an `aside.notes` element.
- The deck opens at the slide named by the URL fragment (`#<data-id>`).

Theme files: a deck at `sections/<dir>/deck/index.html` may reference `../../../theme/`. Recording tools serve the course root, not just the deck directory, so shared files don't need copying into every deck. This depends on support in the recording tool; see PLAN.

## Quiz

```yaml
id: designing-endpoints-quiz
title: Designing Endpoints Quiz
passPercentage: 70
allowRetry: true
maxAttempts: 0               # 0 = unlimited
showFeedback: true
questions:
  - id: q1-idempotent-methods
    type: multiple_choice
    lesson: methods
    prompt: Which methods are idempotent?
    options:
      - {id: a, text: GET, correct: true}
      - {id: b, text: POST, correct: false}
      - {id: c, text: PUT, correct: true}
    explanation: GET and PUT can be repeated with the same result; POST can't.
    points: 1
  - id: q2-404
    type: true_false
    prompt: A 404 means the server failed.
    correctAnswer: false
    explanation: 404 means the resource wasn't found, a client-side condition.
  - id: q3-status-code
    type: short_answer
    prompt: Which status code means "Created"?
    acceptedAnswers: ["201", "201 Created"]
    explanation: 201 Created, with a Location header for the new resource.
```

- Question types are `multiple_choice`, `true_false`, and `short_answer`, matching the AcademyOS grader. The enum values stay snake_case because they are values, not property names.
- `multiple_choice` needs at least two options with unique IDs and at least one correct. With more than one correct, the learner must choose exactly the correct set.
- `true_false` needs a boolean `correctAnswer`.
- `short_answer` needs one or more `acceptedAnswers`, compared ignoring case and repeated spaces.
- `explanation` is required. Feedback without a reason doesn't teach.
- `points` defaults to 1.
- `lesson`, when set, must name a lesson in the same section.

## Published State

`published.json` is written by tools after recording, uploading, or syncing. Authors don't edit it.

```json
{
  "courseVersion": "1.2.0",
  "lessons": {
    "resources": {
      "fingerprint": "sha256:…",
      "durationMs": 151000,
      "videos": {
        "en-US": {"provider": "youtube", "id": "dQw4w9WgXcQ"}
      },
      "captions": {
        "en-US": {"provider": "youtube"}
      }
    }
  },
  "platforms": {
    "academyos": {
      "baseUrl": "https://app.academyos.io",
      "courseId": "…",
      "lessons": {"resources": "…"},
      "quizzes": {"designing-endpoints-quiz": "…"}
    }
  }
}
```

- Keys are spec IDs. Platform IDs are values, so one course can be published to several platforms.
- `fingerprint` is the recording tool's fingerprint of the lesson content the video was made from (videoascode's lesson fingerprint). A tool compares it with the current content to report a lesson as changed after a fresh clone, when `build/` is empty.
- No secrets, tokens, or private URLs belong in this file.

## Validation

`validate.Course(dir)` returns findings. Each finding has a severity (`error` or `warning`), the file, a JSON pointer or line, a message, and a fix. The CLI prints them and exits non-zero on any error.

Checks:

1. Each manifest matches the schema; unknown properties are errors, so typos aren't ignored.
2. IDs: course, section, lesson, and quiz IDs are lowercase kebab-case. Lesson IDs are unique within the course; question IDs within their quiz.
3. Every listed section directory exists and has a manifest; every directory under `sections/` is listed.
4. Each source file exists; each HTML deck parses, its slide IDs are unique, and every `lessons[]` ID is in the deck.
5. Quiz rules from the Quiz section.
6. Every ID in `published.json` still exists or is marked `retired`. A published ID that disappeared is an error naming the likely rename.
7. When `version` is unchanged but `published.json` records a different `courseVersion`, that's a warning: content changed without a version bump.
8. Referenced files (thumbnail, text lessons, `terms.json`) exist.

## JSON Schema

- Generated from the Go types with `invopop/jsonschema` by a `//go:build ignore` generator, one file per document: `course.schema.json`, `section.schema.json`, `quiz.schema.json`, `published.schema.json`.
- Linted with `schemakit lint --property-case camelCase`.
- `$id`: `https://academyos.io/schemas/course-spec/v1/<name>.schema.json`.
- YAML files are checked with the same schema after decoding.

## CLI

```
coursespec validate <dir>            check a course; non-zero exit on errors
coursespec lessons <dir>             list sections, lessons, quizzes, and publish status
coursespec schema [name]             print an embedded schema
```

Output is text by default and `--format json` for tools and agents.

## Compatibility with AcademyOS course.json

AcademyOS's GitHub sync reads a snake_case `course.json` (`schema_version: "1.0"`) that lists lessons inline and leaves quizzes out. AcademyOS keeps reading that file and adds course-spec: a repository with `course.yaml` or `course.json` containing `specVersion` is loaded with this module. The old format is not extended further.
