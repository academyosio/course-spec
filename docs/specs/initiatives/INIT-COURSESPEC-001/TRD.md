# course-spec: Technical Requirements

## Module

- Module path: `github.com/academyosio/course-spec`
- Go package: `coursespec` at the module root, holding the types and the loader.
- Subpackages:

  | Package | Purpose |
  | --- | --- |
  | `coursespec` | Types for every file, and `Load(dir)` which resolves a course directory into one `Course` value |
  | `validate` | Schema and cross-reference checks, returning findings with file, path, message, and fix |
  | `deck` | Discovers slides and lessons in an HTML deck (`data-id`, `data-lesson`, `data-vo`) and computes each lesson's content fingerprint |
  | `schema` | Generated JSON Schema files, embedded with `//go:embed` |
  | `cmd/coursespec` | Cobra CLI: `validate`, `lessons`, `schema` |

- Dependencies are kept small so platforms can import the module: `invopop/jsonschema` (generation only) and `golang.org/x/net/html` for deck parsing. Recording tools import `deck` for lesson discovery, not the other way round, so this module never pulls in browser automation.

## File Format

Every file is JSON with camelCase property names and camelCase enum values. JSON is the one format: agents write it reliably, there is no indentation or implicit typing to get wrong, and one decoder and one schema cover every file. Long text goes in Markdown files that manifests point to, such as `descriptionFile` and text lessons, rather than in long JSON strings.

The loader reads each file's bytes and decodes them with `encoding/json` into the Go types, rejecting unknown properties. Decoding is kept in one function so another input syntax could be added later by converting it to JSON first, without changing the types, the schema, or the validator.

## Directory Layout

```
<course>/
  course.json               course manifest
  description.md            long course description (optional)
  published.json            published state, written by tools, committed
  pronunciations.json       how narration says terms (optional)
  theme/                    shared deck CSS, JS, fonts (optional; see `theme`)
  assets/                   thumbnail, certificate template, course images (optional)
  sections/
    01-<slug>/
      section.json          section manifest
      deck/index.html       media source: slides with narration
      deck/voiceover/transcript.json   narration in every language (optional, committed)
      quiz.json             quiz with answer key (optional)
      text/*.md             text lessons (optional)
      build/                generated: clips, videos, captions, lessons.json (not committed)
```

A course directory is any directory that contains a course manifest. It need not be the repository root, so one repository can hold several courses:

```
courses/
  shared/theme/
  intro-course/course.json
  advanced-course/course.json
```

Rules:

- Section directories are listed in the course manifest in order. A directory under `sections/` that the manifest doesn't list is an error, so a forgotten section can't go unnoticed.
- A section listed as not yet released may have no directory; see Course Manifest.
- The `NN-` prefix is for people browsing the tree; order comes from the manifest. IDs come from the manifests, not from directory names.
- `build/` holds generated files only and belongs in `.gitignore`.

## Course Manifest

```json
{
  "specVersion": "1",
  "id": "building-web-apis",
  "version": "1.2.0",
  "title": "Building Web APIs",
  "summary": "Design, build, and operate an HTTP API.",
  "descriptionFile": "description.md",
  "languages": ["en-US", "es-ES"],
  "audience": "Backend developers who have built a web app.",
  "outcomes": [
    "Design resource-oriented endpoints",
    "Return errors clients can act on"
  ],
  "prerequisites": [],
  "access": "open",
  "tags": ["apis", "http"],
  "authors": [{"name": "Jane Author", "url": "https://example.com"}],
  "thumbnail": "assets/thumbnail.png",
  "theme": "../shared/theme",
  "certificate": {
    "enabled": true,
    "requires": "allQuizzes",
    "title": "API Designer",
    "template": "assets/certificate.html"
  },
  "video": {
    "provider": "youtube",
    "visibility": "public",
    "playlist": {
      "en-US": "Building Web APIs",
      "es-ES": "Creación de APIs web"
    },
    "descriptionFooter": {
      "en-US": "Take the full course with quizzes and a certificate: https://example.com/courses/building-web-apis",
      "es-ES": "Haz el curso completo con cuestionarios y certificado: https://example.com/es/courses/building-web-apis"
    }
  },
  "sections": [
    {"path": "sections/01-designing-endpoints"},
    {"path": "sections/02-handling-errors"},
    {"path": "sections/03-versioning", "released": false, "id": "versioning", "title": "Versioning"}
  ],
  "changes": [
    {
      "version": "1.2.0",
      "notes": "Rewrote the pagination lesson.",
      "retakeQuizzes": ["designing-endpoints-quiz"]
    }
  ]
}
```

- `specVersion` is required and selects the schema. A loader refuses a major version it doesn't know.
- `version` is semantic: major for restructuring, minor for new or rewritten lessons, patch for corrections.
- `description` holds a short Markdown description inline; `descriptionFile` points to a Markdown file for a long one. Set one, not both.
- `languages` lists the languages the course is published in; the first is the default.
- `prerequisites` lists course IDs.
- `access` is `open`, `invite`, or `paid`.
- `theme` is a path to a directory inside the same repository, defaulting to `theme/` in the course directory. It lets several courses share one theme without copies.
- `certificate.requires` is `allQuizzes`, `allLessons`, or `allLessonsAndQuizzes`. `title` defaults to the course title; `template` defaults to the platform's template.
- `video` is read by upload tools: `provider` is `youtube` or `vimeo`, and `visibility` is `public`, `unlisted`, or `private`. `playlist` and `descriptionFooter` are keyed by language, because each language gets its own playlist; a language without an entry has no playlist or footer. The spec records how videos are published; uploading is the tool's job.
- `changes[].retakeQuizzes` names quizzes whose earlier passes a platform should treat as out of date.
- `sections` lists each section as an object with its `path`. Every entry is an object, so the schema has no union type.
- A section with `"released": false` is planned but not published, so a course can show its whole outline while it is released a section at a time. It carries its own `id` and `title`, and its directory may be missing or incomplete. Validation checks only that its ID is unique and kebab-case. A platform shows it as coming soon, without lessons or a quiz, and issues no certificates while any section is unreleased, even when `certificate.enabled` is true.

## Section Manifest

```json
{
  "id": "designing-endpoints",
  "title": "Designing Endpoints",
  "summary": "Resources, methods, and status codes.",
  "source": {"type": "htmlDeck", "path": "deck/index.html"},
  "lessons": [
    {
      "id": "resources",
      "objectives": ["Model a domain as resources"],
      "free": true
    }
  ],
  "quiz": "quiz.json"
}
```

- `lessons` adds details to discovered lessons by ID, or lists lessons for sources that don't define them.
- `quiz` defaults to `quiz.json` when that file exists.
- A section has at least one lesson or a quiz. A section with `source.type` `none`, no lessons, and a quiz is a quiz-only section, used for a final assessment: its questions may point at lessons in any section, and a platform shows it as one quiz after the previous section. On AcademyOS, where a quiz belongs to a lesson, the import creates one quiz lesson for it.

Sources:

| `source.type` | Lessons come from | Video comes from |
| --- | --- | --- |
| `htmlDeck` | Slides with `data-lesson`; the lesson ID is the starting slide's `data-id` | Recorded from the deck |
| `marp` | Slides marked as lesson starts (convention defined when supported) | Recorded from the deck |
| `video` | `lessons[]`, each with `file` (relative path, not committed) or an already-published video | The file |
| `none` | `lessons[]`, text lessons only, or none in a quiz-only section | None |

`marp` and `video` are reserved for Phase 3; until then the validator reports them as not yet supported.

In an `htmlDeck` section, a `lessons[]` entry with `type` `text` is a text lesson placed after the lesson named by `after`, or at the end of the section without it.

Lesson fields:

| Field | Meaning |
| --- | --- |
| `id` | Required. Unique within the course |
| `title` | Required when not discovered; overrides a discovered title |
| `type` | `video` (default with a media source) or `text` |
| `content` | Path to a Markdown file, for text lessons, such as `text/glossary.md` |
| `objectives` | Learning objectives |
| `free` | Free preview |
| `estimatedMinutes` | Used until a recording gives the real duration |
| `after` | For a text lesson, the ID of the lesson it follows |
| `retired` | Kept to preserve a published ID; not shown to new learners |

An entry in `lessons[]` that isn't a text lesson and whose ID the deck doesn't contain is an error, so a renamed slide is caught.

## HTML Deck Conventions

The spec adopts the videoascode conventions so the same deck records and validates without changes:

- A slide is an element with class `slide` and a `data-id`, unique in the deck.
- A slide with `data-lesson` starts a lesson. The value is the title; when empty, the slide's `data-title` or ID is used.
- Slides before the first lesson start belong to the first lesson. A deck with no lesson starts is one lesson, with the section's ID.
- Narration in the default language comes from the deck: a slide's `data-vo`, or its `aside.notes` element. The deck is the only source for that language.
- The deck opens at the slide named by the URL fragment (`#<data-id>`).
- Narration may contain `[PAUSE:ms]` directives, which are timing for the recording tool and not part of the spoken text.

A deck is self-contained: every file it loads (styles, scripts, fonts, images, media) is in the course directory or its theme, referenced by a relative path. Players that sandbox decks, and offline playback, block network requests, so a stylesheet from a CDN or a web font from a font service fails there. The validator checks that each relative reference resolves and warns about each remote one. Links a learner clicks (`<a href>`) are not loads and may point anywhere.

Theme files: the deck references the course's `theme` directory with a relative path, which the validator checks resolves. Recording tools serve the nearest directory that contains both the course and its theme, not just the deck directory, so shared files don't need copying into every deck. This depends on support in the recording tool; see PLAN.

### Translations

Translations are in `deck/voiceover/transcript.json`, in the videoascode transcript format, defined by its JSON Schema (`pkg/transcript/transcript.schema.json` in videoascode). The file lists slides; each slide has its narration per language; each language has a list of segments, one per sentence. Translations are written by hand, so the file is committed source, not build output. Without it, the deck has only the default language.

- Each slide entry carries `id`, equal to the slide's `data-id`, and translations are matched to slides by `id`. The entry's `index` is informational. Matching by position would attach every later translation to the wrong slide after a slide is inserted or deleted, and nothing would report it.
- A transcript written before slide entries carried `id` is matched by position, with a warning to add the IDs.
- Default-language entries in the transcript are generated from the deck. Tools use the deck when the two differ, and the validator warns.
- Transcript `metadata`, such as the default voice, is a recording setting. course-spec doesn't read it.

### Content Fingerprint

`deck` computes a fingerprint for each lesson, written `sha256:<hex>`. For each slide in the lesson, in order, it hashes:

1. the slide's HTML;
2. for each language in the course's `languages`, sorted by code, the slide's narration: from the deck for the default language, from the transcript for the others, with the segment texts joined by single spaces. A language with no narration for the slide contributes empty text.

Narration is normalized before hashing: `[PAUSE:ms]` directives are removed and runs of whitespace become one space. Each value is written with its length in bytes before it, so field boundaries can't shift between slides.

Joining the segments means splitting the same sentences differently isn't a content change. Transcript `metadata` is never hashed. The fingerprint covers everything an author writes and nothing a recording tool chooses: the voice, pronunciations, timing, and resolution belong in the recording tool's own fingerprint.

The fingerprint doesn't cover the theme's CSS, JavaScript, fonts, or images, so a theme change doesn't mark any lesson as changed. After changing a theme, re-record the affected lessons by ID, or all of them. A theme fingerprint in `published.json` can be added if this proves a problem.

## Pronunciations

`pronunciations.json` maps a term to how TTS should say it, per language. It uses the videoascode pronunciation map:

```json
{
  "SaaS": {"en-US": "sass"},
  "OAuth": {"en-US": "oh auth"},
  "API": {"en-US": "A P I"}
}
```

Terms match case-insensitively at word boundaries. Only the TTS input changes; captions keep the written term.

## Quiz

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

- Question types are `multipleChoice`, `trueFalse`, and `shortAnswer`. A platform maps them to its own names on import.
- `multipleChoice` needs at least two options with unique IDs and at least one correct. With more than one correct, the learner must choose exactly the correct set.
- `trueFalse` needs a boolean `correctAnswer`.
- `shortAnswer` needs one or more `acceptedAnswers`, compared ignoring case and repeated spaces.
- `explanation` is required. Feedback without a reason doesn't teach.
- `points` defaults to 1. `maxAttempts` of 0 means unlimited.
- `lesson`, when set, must name a lesson in the course. A review quiz at the end of a course can point each question at the lesson it covers, in any section.
- `shuffleOptions` (default false) asks the platform to shuffle multiple-choice options for each learner. Without it, options are shown in file order, so the validator warns when correct options keep landing in the same position.

## Published State

`published.json` is written by tools after recording, uploading, or syncing. Authors don't edit it.

```json
{
  "courseVersion": "1.2.0",
  "lessons": {
    "resources": {
      "contentFingerprint": "sha256:…",
      "recordingFingerprint": "3f9a1c0b7d2e4a65",
      "durationMs": 151000,
      "videos": {
        "en-US": {"provider": "youtube", "id": "dQw4w9WgXcQ"}
      },
      "captions": {
        "en-US": {"provider": "youtube"}
      }
    }
  },
  "quizzes": {
    "designing-endpoints-quiz": {"fingerprint": "sha256:…"}
  },
  "video": {
    "playlists": {"en-US": {"provider": "youtube", "id": "PL…"}}
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
- `contentFingerprint` is the `deck` content fingerprint the video was made from. `coursespec lessons` compares it with the current content to report a lesson as **changed**, even after a fresh clone when `build/` is empty.
- `recordingFingerprint` is the recording tool's own fingerprint, an opaque string in whatever format that tool uses. The `sha256:` form applies only to course-spec's content and quiz fingerprints. When only it differs (a new voice, say), the recording tool reports the lesson as needing a re-record; course-spec doesn't interpret it.
- `quizzes[].fingerprint` hashes a quiz's questions and settings, so a changed quiz is noticed the same way.
- No secrets, tokens, or private URLs belong in this file.

## Validation

`validate.Course(dir)` returns findings. Each finding has a severity (`error` or `warning`), the file, a JSON pointer or line, a message, and a fix. The CLI prints them and exits non-zero on any error.

Checks:

1. Each manifest matches the schema; unknown properties are errors, so typos aren't ignored.
2. IDs: course, section, lesson, and quiz IDs are lowercase kebab-case. Lesson IDs are unique within the course; question IDs within their quiz.
3. Every listed section directory exists and has a manifest, except unreleased sections; every directory under `sections/` is listed. Every section has a lesson or a quiz.
4. Each source file exists; each HTML deck parses, its slide IDs are unique, and every `lessons[]` ID that isn't a text lesson is in the deck. Each relative file a deck loads exists.
5. Quiz rules from the Quiz section.
6. Every ID in `published.json` still exists or is marked `retired`. A published ID that disappeared is an error naming the likely rename.
7. When a lesson's content fingerprint or a quiz's fingerprint differs from `published.json`, and `version` equals the published `courseVersion`, warn: content changed without a version bump.
8. Referenced files (thumbnail, certificate template, text lessons, `pronunciations.json`, transcripts) exist. `theme` resolves inside the repository; a path that leaves it is an error.
9. Transcripts: a slide `id` that isn't in the deck is an error; it catches a translation left behind by a renamed or deleted slide. A deck slide with no narration in a language listed in `languages` is a warning, because recording tools record a language only when every slide has it.

Warnings that help authors and agents:

- Without `shuffleOptions`, three or more multiple-choice questions in a quiz whose correct options all sit at the same position.
- A slide with no narration (`data-vo` or `aside.notes`).
- A deck that loads a file from another host, which sandboxed and offline players block.
- A lesson whose narration runs over about 10 minutes, estimated at 150 words per minute.
- Narration containing characters TTS reads badly, such as `(`, `)`, `/`, or `→`. Off by default; enabled with `--strict-narration`.

## JSON Schema

- Generated from the Go types with `invopop/jsonschema` by a `//go:build ignore` generator, one file per document: `course.schema.json`, `section.schema.json`, `quiz.schema.json`, `published.schema.json`.
- Linted with `schemakit lint --property-case camelCase`.
- `$id`: `https://academyos.io/schemas/course-spec/v1/<name>.schema.json`.

## CLI

```
coursespec validate <dir>            check a course; non-zero exit on errors
coursespec lessons <dir>             list sections, lessons, quizzes, and publish status
coursespec schema [name]             print an embedded schema
```

`coursespec lessons` compares content fingerprints, so it doesn't see theme changes; see Content Fingerprint.

Output is text by default and `--format json` for tools and agents.

Given a directory without a course manifest, `validate` and `lessons` act on every course directory beneath it, skipping `.git`, `node_modules`, and `build/` directories.

## AcademyOS Course Sync

AcademyOS's GitHub sync currently reads its own snake_case course file, which lists lessons inline and leaves quizzes out. AcademyOS replaces that format with course-spec rather than supporting both: its sync loads a course directory with this module, and the old format is retired. No published course depends on it.
