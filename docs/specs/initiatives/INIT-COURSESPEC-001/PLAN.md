# course-spec: Plan

## Approach

Write the spec from the terms already in use, and prove it against a real course before calling it done.

The terms are not new. Course, section, lesson, slide, clip, and chapter are already implemented in videoascode's HTML deck support, and quizzes in the AcademyOS grader. A course built without the spec has used the same conventions: lesson IDs from `data-id`, one deck per section, one quiz per section in the grader's question format. So the spec starts from those, and a real course is ported to it while the types are written, not after. Every gap the port finds becomes a change to the spec.

## Steps

1. **Types and loader.** Go types for the course, section, quiz, and published-state files; `Load(dir)` resolves a directory into one value, discovering lessons from HTML decks.
2. **Deck discovery and content fingerprints.** The `deck` package finds slides and lessons and hashes each lesson's slides and narration in every language.
3. **Schema.** Generate and embed the JSON Schema; lint it.
4. **Validator and CLI.** `coursespec validate` and `coursespec lessons`, covering every check and warning in the TRD.
5. **Port a real course.** Convert an existing multi-section course with decks and quizzes:

   - the course manifest and its generated lesson list become a course-spec `course.json` plus one `section.json` per section, with lessons discovered from the decks;
   - each quiz moves into its section as `quiz.json`, with property names and question types in camelCase;
   - a theme copied into every deck becomes one shared `theme` directory;
   - empty video placeholders are dropped; recording and upload write `published.json`;
   - the course's own checker scripts are replaced by `coursespec validate`.

6. **Example course.** A small course in `examples/` covering every source type and question type, used by tests and docs.
7. **Docs site.** MkDocs: terms, layout, each file's reference, validation errors, and a guide for AI agents writing a course.
8. **Release v0.1.0.**

## Dependent Work in Other Projects

| Project | Change |
| --- | --- |
| AcademyOS | Replace its own course sync format with course-spec: load course directories in GitHub sync, import quizzes, honor `shuffleOptions`, keep platform IDs stable across syncs, map `version` and `changes` to course versions, use the certificate title and template, and write its IDs back to `published.json` |
| videoascode | Import `deck` from this module for lesson discovery and build its recording fingerprint on the content fingerprint; write each transcript slide's `id` and match translations by `id`, falling back to position with a warning for older files; record a whole course from `course.json` (`vac course video <dir> --lessons changed`); write build output to each section's `build/`; write fingerprints, durations, and uploaded videos to `published.json`; serve the directory holding the course and its theme |
| H5pKit | Convert a course-spec quiz to H5P content |

The import direction is deliberate: `deck` depends only on an HTML parser, while videoascode's deck package brings in browser automation. videoascode imports course-spec, never the reverse.

## Open Questions

1. **Quiz placement.** Quizzes come after a section's lessons. A course-level final exam is RMI-COURSESPEC-015; since questions can already reference any lesson, it needs only a place in the manifest.
2. **License.** MIT for the code and CC BY 4.0 for the spec text is the usual split; MIT for everything is simpler. Decide before the docs site is published (RMI-COURSESPEC-008).
3. **QTI export.** Worth adding once a platform other than AcademyOS needs the quizzes.
