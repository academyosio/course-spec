# The Course Directory

A course directory is any directory that contains a `course.json`. It need
not be a repository's root, so one repository can hold several courses and
share a theme between them:

```text
courses/
  shared/theme/
  intro-course/course.json
  advanced-course/course.json
```

Every file is JSON with camelCase property names and camelCase enum values.
Unknown properties are errors, so a misspelled property is reported rather
than ignored. Long text, such as a course description or a text lesson, goes
in Markdown files the manifests point to.

## `course.json`

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
  "outcomes": ["Design resource-oriented endpoints"],
  "prerequisites": [],
  "access": "open",
  "tags": ["apis", "http"],
  "authors": [{"name": "Jane Author", "url": "https://example.com"}],
  "thumbnail": "assets/thumbnail.png",
  "theme": "../shared/theme",
  "certificate": {"enabled": true, "requires": "allQuizzes", "title": "API Designer"},
  "video": {"provider": "youtube", "visibility": "public"},
  "sections": [
    {"path": "sections/01-designing-endpoints"},
    {"path": "sections/02-handling-errors"},
    {"path": "sections/03-versioning", "released": false, "id": "versioning", "title": "Versioning"}
  ],
  "changes": [
    {"version": "1.2.0", "notes": "Rewrote the pagination lesson.", "retakeQuizzes": ["designing-endpoints-quiz"]}
  ]
}
```

| Property | Meaning |
| --- | --- |
| `specVersion` | Required. `"1"`, or `"1.<minor>"`. A reader refuses a major version it doesn't know |
| `id` | Required. Lowercase kebab-case; never changes once published |
| `version` | Required. Semantic: major for restructuring, minor for new or rewritten lessons, patch for corrections |
| `title` | Required |
| `summary` | One line |
| `description` / `descriptionFile` | A short Markdown description, or a Markdown file for a long one. Set one, not both |
| `languages` | The languages the course is published in; the first is the default. Absent means `en-US` |
| `audience`, `outcomes`, `prerequisites`, `tags` | `prerequisites` lists course IDs |
| `access` | `open`, `invite`, or `paid` |
| `authors` | Names and optional URLs |
| `thumbnail` | An image in the course directory |
| `theme` | A directory of shared deck files inside the same repository; default `theme/` |
| `certificate` | `enabled`; `requires` is `allQuizzes`, `allLessons`, or `allLessonsAndQuizzes`; `title` defaults to the course title; `template` defaults to the platform's |
| `video` | How videos are published: `provider` (`youtube` or `vimeo`), `visibility` (`public`, `unlisted`, `private`), and per-language `playlist` titles and `descriptionFooter` text |
| `sections` | Required. Each section as `{"path": …}`, in order |
| `changes` | The change log; `retakeQuizzes` names quizzes whose earlier passes are out of date |

### Sections

Each entry in `sections` is an object with the section directory's `path`.
The `NN-` prefix on directory names is for people browsing the tree; the
order comes from this list, and IDs come from the manifests.

Every directory under `sections/` must be listed, so a forgotten section
can't go unnoticed.

A section listed with `"released": false` is planned but not published. It
carries its own `id` and `title`, and its directory may be missing or
incomplete, so a course can show its whole outline while it is released a
section at a time. A platform shows it as coming soon and issues no
certificates while any section is unreleased.

## `section.json`

```json
{
  "id": "designing-endpoints",
  "title": "Designing Endpoints",
  "summary": "Resources, methods, and status codes.",
  "source": {"type": "htmlDeck", "path": "deck/index.html"},
  "lessons": [
    {"id": "resources", "objectives": ["Model a domain as resources"], "free": true},
    {"id": "glossary", "type": "text", "title": "Glossary", "content": "text/glossary.md", "after": "resources"}
  ],
  "quiz": "quiz.json"
}
```

| Property | Meaning |
| --- | --- |
| `id`, `title` | Required. The ID is kebab-case and unique among sections |
| `summary` | One line |
| `source` | Where lessons come from: `htmlDeck` with a `path`, or `none` |
| `lessons` | Details for lessons discovered from the deck, matched by ID, or the lessons themselves for a `none` source |
| `quiz` | The quiz file; defaults to `quiz.json` when that exists |

`marp` and `video` sources are reserved for a later version and reported as
not yet supported.

### Lessons

With an `htmlDeck` source, lessons come from the deck: each slide with
`data-lesson` starts one, and its `data-id` is the lesson's ID. A `lessons`
entry adds details to a discovered lesson by its ID:

| Property | Meaning |
| --- | --- |
| `id` | Required. Unique within the course |
| `title` | Overrides the discovered title; required for lessons not in the deck |
| `type` | `video` (the default with a deck) or `text` |
| `content` | The Markdown file of a text lesson |
| `objectives` | Learning objectives |
| `free` | A free preview |
| `estimatedMinutes` | Used until a recording gives the real length |
| `after` | Places a text lesson after the lesson with this ID; without it, at the end of the section |
| `retired` | Kept to preserve a published ID; not shown to new learners |

An entry that isn't a text lesson and whose ID the deck doesn't contain is an
error, unless it's retired. That catches a renamed slide.

### Quiz-only sections

A section with `source.type` `none`, no lessons, and a quiz is a quiz-only
section, used for a final assessment. Its questions may point at lessons in
any section. Every section needs at least one lesson or a quiz.

```json
{
  "id": "final-assessment",
  "title": "Final Assessment",
  "source": {"type": "none"}
}
```

## Other files

| File | Contents |
| --- | --- |
| `description.md` | The long description, when `descriptionFile` names it |
| `pronunciations.json` | How narration says terms; see [Decks and narration](decks-and-narration.md#pronunciations) |
| `published.json` | Written by tools; see [Published state](published-state.md) |
| `theme/` | Shared deck styles, scripts, fonts, and images |
| `sections/*/build/` | Generated clips, videos, and captions; belongs in `.gitignore` |
