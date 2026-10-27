# Writing a Course with an AI Agent

course-spec is designed for agents: every file is JSON with a published
schema, lessons come from the slides instead of being repeated, and the
validator's findings carry a code and a fix an agent can act on.

## Give the agent these instructions

Put this in the course repository's `CLAUDE.md` or `AGENTS.md`:

```markdown
## Course format

Courses are course-spec directories (https://github.com/academyosio/course-spec).

- After every change, run `coursespec validate --format json <dir>` and fix
  every error before finishing. Fix warnings unless there is a reason not to.
- Lessons come from the deck: a slide with `data-lesson` starts a lesson,
  and its `data-id` is the lesson ID. Don't list them again in section.json
  except to add details.
- Never change a published lesson ID (a `data-id` with `data-lesson`), a
  quiz ID, or a question ID. To remove a published lesson, keep it in
  section.json with `"retired": true`.
- Narration for the default language goes in each slide's `data-vo`; other
  languages go in deck/voiceover/transcript.json, with each slide entry's
  `id` equal to the slide's data-id.
- Decks load only local files by relative paths. No CDNs or web fonts.
- Every quiz question needs an explanation. Vary where correct options sit,
  or set `shuffleOptions`.
- Raise `version` in course.json when content changes: minor for new or
  rewritten lessons, patch for corrections.
- Never edit published.json; tools write it.
```

## Give it the schemas

The JSON Schemas tell an agent every property and enum value. Print them
with `coursespec schema <name>`, or point the agent at
`https://academyos.io/schemas/course-spec/v1/<name>.schema.json` once
published there. The schema files are in the repository's `schema/`
directory.

## Give it an example

The repository's
[`testdata/courses/web-apis`](https://github.com/academyosio/course-spec/tree/main/testdata/courses/web-apis)
uses every feature of version 1: a deck with lessons and a title slide, a
text lesson placed after a deck lesson, a translation, pronunciations, a
quiz with all three question types, a quiz-only final assessment, and a
planned section. An agent that copies its shape rarely needs more than one
round of fixes.

## The loop

1. The agent writes or edits files.
2. It runs `coursespec validate --format json`.
3. For each finding, it reads `code`, `file`, `pointer`, and `fix`, and
   changes that place.
4. It repeats until there are no errors, then reviews the warnings.

Common first-round findings:

| Code | Usual cause |
| --- | --- |
| `invalid-json` | A snake_case or misspelled property |
| `lesson-not-in-deck` | A lesson listed in section.json under a different ID than its slide |
| `missing-slide-id` | Slides written without `data-id` |
| `missing-explanation` | Quiz questions without explanations |
| `unlisted-section` | A new section directory not added to course.json |
| `remote-deck-file` | A stylesheet or font from a CDN |
