# course-spec

course-spec is an open format for keeping a complete online course in one
directory under version control: the outline, HTML slide decks with
narration, translations, quizzes with answer keys, and a record of what has
been published. People and AI agents write it by hand; tools check it,
record it to video, and import it into a learning platform.

```text
building-web-apis/
  course.json                    the course: title, version, languages, sections
  pronunciations.json            how narration says terms
  published.json                 what tools have published, written by them
  theme/deck.css                 shared deck styles
  sections/
    01-designing-endpoints/
      section.json               the section and its lessons
      deck/index.html            slides with narration; data-lesson starts a lesson
      deck/voiceover/transcript.json   narration in other languages
      quiz.json                  the section's quiz with its answer key
```

## Why a format

Course platforms keep a course in their own database, edited through a web
form. That works for one author clicking through a UI, but not for a course
that is written as code and reviewed in pull requests, written partly by AI
agents, recorded to video by tools, or published to more than one place.
course-spec keeps the whole course in files, with:

- **Stable IDs** for every course, section, lesson, quiz, and question, so
  learner progress and published videos survive edits.
- **Lessons discovered from decks,** so the outline can't drift from the
  slides.
- **Quizzes as part of the course,** in one documented format.
- **Source, build output, and published state kept apart,** so a fresh clone
  knows what is live.
- **A validator** whose findings name the file, the place, and the fix.

## What's here

| | |
| --- | --- |
| [Getting started](guides/getting-started.md) | Install `coursespec` and check your first course |
| [The course directory](guides/course-directory.md) | `course.json`, `section.json`, sections, and lessons |
| [Decks and narration](guides/decks-and-narration.md) | Slides, lessons, narration, translations, pronunciations, and self-contained decks |
| [Quizzes](guides/quizzes.md) | Question types and the rules a quiz must meet |
| [Published state](guides/published-state.md) | `published.json` and content fingerprints |
| [Validating a course](guides/validating.md) | `coursespec validate` and `coursespec lessons` |
| [AI agents](guides/ai-agents.md) | Instructions to give an agent writing a course |
| [Go library](guides/library.md) | Load and check courses from Go |
| [Validation findings](reference/findings.md) | Every finding code |
| [JSON Schemas](reference/schemas.md) | The schema for each file |

## Tools that use it

- **[videoascode](https://github.com/grokify/videoascode)** records HTML
  decks to video; course-spec adopts its deck and transcript conventions.
- **[AcademyOS](https://academyos.io)** imports course-spec directories,
  plays their decks with browser speech, and grades their quizzes.

The format is licensed MIT, like the code, schemas, and examples.
