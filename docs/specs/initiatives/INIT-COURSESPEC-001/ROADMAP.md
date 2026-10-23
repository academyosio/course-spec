# ROADMAP — Course directory format and validator

**Initiative:** `INIT-COURSESPEC-001`
**Repository:** `github.com/academyosio/course-spec`

Specs: [PRD](PRD.md), [TRD](TRD.md), [PLAN](PLAN.md). Work items carry
stable IDs (`RMI-COURSESPEC-NNN`), are reviewed and executed by phase, and
commits implementing one carry the trailer `Refs: RMI-COURSESPEC-<NNN>`.

## Phase 1 — Format and Validator

**Theme:** A course directory loads into one value and is checked by one command.

- [ ] `RMI-COURSESPEC-001` Go types for course, section, lesson, quiz, and published state
- [ ] `RMI-COURSESPEC-002` Course directory loader with YAML and JSON support
  - Depends on: `RMI-COURSESPEC-001`
- [ ] `RMI-COURSESPEC-003` HTML deck lesson discovery
- [ ] `RMI-COURSESPEC-004` Generated, embedded, and linted JSON Schema
  - Depends on: `RMI-COURSESPEC-001`
- [ ] `RMI-COURSESPEC-005` `coursespec validate` and `coursespec lessons` with text and JSON output
  - Depends on: `RMI-COURSESPEC-002`, `RMI-COURSESPEC-003`, `RMI-COURSESPEC-004`

## Phase 2 — Proven and Documented

**Theme:** A real course validates, and people and agents can learn the format from its docs.

- [ ] `RMI-COURSESPEC-006` Port a real multi-section course and fold every gap back into the spec
  - Depends on: `RMI-COURSESPEC-005`
- [ ] `RMI-COURSESPEC-007` Example course covering every source type and question type
  - Depends on: `RMI-COURSESPEC-005`
- [ ] `RMI-COURSESPEC-008` MkDocs site: terms, layout, file reference, validation errors
- [ ] `RMI-COURSESPEC-009` Guide for AI agents writing a course
  - Depends on: `RMI-COURSESPEC-007`
- [ ] `RMI-COURSESPEC-010` v0.1.0 release
  - Depends on: `RMI-COURSESPEC-006`, `RMI-COURSESPEC-008`

## Phase 3 — More Sources and Exports

**Theme:** Courses not made from HTML decks fit, and quizzes reach other platforms.

- [ ] `RMI-COURSESPEC-011` Marp deck source
- [ ] `RMI-COURSESPEC-012` Video file source
- [ ] `RMI-COURSESPEC-013` Published-state writer helpers for recording and platform tools
- [ ] `RMI-COURSESPEC-014` QTI export for quizzes
- [ ] `RMI-COURSESPEC-015` Course-level final exam
