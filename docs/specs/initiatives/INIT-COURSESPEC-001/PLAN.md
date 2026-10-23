# course-spec: Plan

## Approach

Write the spec from the terms already in use, and prove it against a real course before calling it done.

The terms are not new. Course, section, lesson, slide, clip, and chapter are already implemented in videoascode's HTML deck support, and quizzes in the AcademyOS grader. A course built without the spec has used the same conventions: lesson IDs from `data-id`, one deck per section, one quiz per section in the grader's question format. So the spec starts from those, and a real course is ported to it while the types are written, not after. Every gap the port finds becomes a change to the spec.

## Steps

1. **Types and loader.** Go types for the course, section, quiz, and published-state files; `Load(dir)` resolves a directory into one value, discovering lessons from HTML decks.
2. **Schema.** Generate and embed the JSON Schema; lint it.
3. **Validator and CLI.** `coursespec validate` and `coursespec lessons`, covering every check in the TRD.
4. **Port a real course.** Convert an existing multi-section course with decks and quizzes:

   - the course manifest and its generated lesson list become `course.yaml` plus one `section.yaml` per section, with lessons discovered from the decks;
   - each quiz moves into its section as `quiz.yaml`, with property names in camelCase;
   - empty video placeholders are dropped; recording and upload write `published.json`;
   - the course's own checker scripts are replaced by `coursespec validate`.

5. **Example course.** A small course in `examples/` covering every source type and question type, used by tests and docs.
6. **Docs site.** MkDocs: terms, layout, each file's reference, validation errors, and a guide for AI agents writing a course.
7. **Release v0.1.0.**

## Dependent Work in Other Projects

| Project | Change |
| --- | --- |
| AcademyOS | Load course-spec directories in GitHub sync, import quizzes, keep platform IDs stable across syncs, map `version` and `changes` to course versions, and write its IDs back to `published.json` |
| videoascode | Record a whole course from `course.yaml` (`vac course video <dir> --lessons changed`), write build output to each section's `build/`, write fingerprints and durations to `published.json`, and serve the course root so decks share `theme/` |
| H5pKit | Convert a course-spec quiz to H5P content |

## Open Questions

1. **Deck parser ownership.** The `deck` package duplicates videoascode's lesson discovery. videoascode could import it, keeping one implementation.
2. **Quiz placement.** Quizzes come after a section's lessons. Quizzes placed mid-section, or a final exam at course level, may be needed later.
3. **License.** MIT for the code. The spec text could be CC BY 4.0 instead.
4. **QTI export.** Worth adding once a platform other than AcademyOS needs the quizzes.
