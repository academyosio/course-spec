# course-spec

![coverage](https://img.shields.io/badge/coverage-80.6%25-green)

An open format for keeping a complete online course in one directory under
version control: the outline, HTML slide decks with narration, translations,
quizzes with answer keys, and a record of what has been published. People and
AI agents write it by hand; tools check it, record it to video, and import it
into a learning platform.

This repository holds the Go types, a loader, a validator, generated JSON
Schemas, and the `coursespec` command.

## Install

```bash
go install github.com/academyosio/course-spec/cmd/coursespec@latest
```

## Commands

```bash
coursespec validate <dir>      # check a course; exits non-zero on errors
coursespec lessons <dir>       # list sections, lessons, and quizzes with publish status
coursespec schema [name]       # print a JSON Schema: course, section, quiz, published
```

Given a directory without a `course.json`, `validate` and `lessons` act on
every course beneath it. Add `--format json` for tools and agents.

Each finding names the file, a JSON pointer or slide, the problem, and a fix:

```text
error: sections/01-designing-endpoints/section.json /lessons/2/id: lesson "pagination" is not in the deck [lesson-not-in-deck]
  fix: use the data-id of a slide with data-lesson, or set "type": "text" for a text lesson
```

## A Course Directory

```text
<course>/
  course.json                    course manifest
  pronunciations.json            how narration says terms (optional)
  published.json                 written by tools after publishing
  theme/                         shared deck CSS, JS, fonts (optional)
  sections/
    01-<slug>/
      section.json               section manifest
      deck/index.html            slides with narration
      deck/voiceover/transcript.json   translations (optional)
      quiz.json                  quiz with answer key (optional)
```

Every file is JSON with camelCase property names. A complete example is in
[`testdata/courses/web-apis`](testdata/courses/web-apis).

### Decks

A deck is an HTML page. Each slide is an element with class `slide` and a
stable `data-id`; narration is in `data-vo` or a child `aside class="notes"`;
a slide with `data-lesson` starts a lesson, whose ID is that slide's
`data-id`. These are the [videoascode](https://github.com/grokify/videoascode)
conventions, so the same deck can be recorded to video.

A deck is self-contained: it loads styles, scripts, fonts, and images from
the course directory or its theme by relative paths. Players that sandbox
decks block remote files, so the validator warns about them.

### Translations

Narration in other languages is in `deck/voiceover/transcript.json`, in the
videoascode transcript format. Each slide entry carries `id`, the slide's
`data-id`, so translations stay with their slides when slides are added or
removed.

### Published State and Fingerprints

Tools write `published.json` after recording or importing a course. For each
lesson it records a content fingerprint: a hash of the lesson's slides and
narration in every language. `coursespec lessons` compares fingerprints to
report what changed since publishing, even in a fresh clone.

## Library

```go
c, err := coursespec.Load("courses/building-web-apis")   // stops at the first problem
findings := validate.Course("courses/building-web-apis", validate.Options{})
report := c.Status()                                       // lessons and quizzes with status
```

| Package | Purpose |
| --- | --- |
| `coursespec` | Types for every file, `Load`, `Find`, and publish status |
| `deck` | Slides, lessons, transcripts, and content fingerprints from HTML decks |
| `validate` | Every check, as findings with a fix |
| `schema` | The generated JSON Schemas, embedded |

The Go types are the source of truth. After changing them, regenerate the
schemas with `go generate ./schema`.

## Documentation

Guides for the format, validation, AI agents, and the Go library are in
[`docs/`](docs/index.md), a MkDocs site: `pip install mkdocs-material`, then
`mkdocs serve`. Start with [Getting started](docs/guides/getting-started.md).

Changes are in [CHANGELOG.md](CHANGELOG.md), generated from `CHANGELOG.json`,
and release notes in [`docs/releases/`](docs/releases/).

## Specification

The requirements, design, plan, and roadmap are in
[`docs/specs/initiatives/INIT-COURSESPEC-001`](docs/specs/initiatives/INIT-COURSESPEC-001).

## License

MIT
