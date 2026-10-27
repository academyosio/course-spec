# Validation Findings

Every finding `coursespec validate` reports, by code. Errors make a course
invalid; warnings don't. Each finding also carries its own fix.

## Files and JSON

| Code | Severity | Meaning |
| --- | --- | --- |
| `missing-course` | error | The directory has no `course.json` |
| `invalid-json` | error | A file doesn't decode: bad JSON, a wrong type, an unknown property, or an unsupported `specVersion` |
| `unreadable` | error | A file or path couldn't be read |
| `missing-file` | error | A referenced file doesn't exist: description, thumbnail, certificate template, deck, quiz, or text lesson |
| `path-outside-course` | error | A path leaves the course directory |
| `missing-path` | error | A section entry or deck source has no `path` |
| `unused-property` | warning | A property that has no effect where it is, such as `file` on a deck lesson |
| `invalid-enum` | error | A value that isn't one of the allowed ones, such as `access`, `source.type`, or a question `type` |

## The course

| Code | Severity | Meaning |
| --- | --- | --- |
| `invalid-id` | error | A course, section, lesson, or quiz ID isn't lowercase kebab-case |
| `invalid-version` | error | `version` or a change's version isn't semantic |
| `missing-title` | error | A course, section, text lesson, unreleased section, or quiz has no title |
| `description-twice` | error | Both `description` and `descriptionFile` are set |
| `duplicate-language` | error | A language is listed twice |
| `unknown-language` | warning | A video playlist or footer names a language the course doesn't list |
| `missing-theme` | error | The theme directory doesn't exist |
| `theme-outside-repository` | error | The theme is outside the repository |
| `no-sections` | error | The course lists no sections |
| `unknown-quiz` | error | `changes[].retakeQuizzes` names a quiz that doesn't exist |
| `unversioned-change` | warning | Content changed since it was published at this `version` |

## Sections and lessons

| Code | Severity | Meaning |
| --- | --- | --- |
| `unlisted-section` | error | A directory under `sections/` isn't in `course.json` |
| `missing-section` | error | A released section has no `section.json` |
| `duplicate-section-id` | error | Two sections share an ID |
| `empty-section` | error | A section has no lessons and no quiz |
| `unsupported-source` | error | `marp` and `video` sources aren't supported yet |
| `lesson-not-in-deck` | error | A `lessons[]` entry that isn't a text lesson names no lesson in the deck |
| `duplicate-lesson-id` | error | Two lessons in the course share an ID |
| `missing-content` | error | A text lesson has no `content` file |
| `video-without-source` | error | A video lesson in a section with a `none` source |
| `unknown-lesson` | error | `after`, or a question's `lesson`, names a lesson that doesn't exist |

## Decks and narration

| Code | Severity | Meaning |
| --- | --- | --- |
| `invalid-deck` | error | The deck has no `class="slide"` elements or doesn't parse |
| `missing-slide-id` | error | Slides without a `data-id` |
| `duplicate-slide-id` | error | Two slides share a `data-id` |
| `missing-deck-file` | error | The deck loads a relative file that doesn't exist |
| `remote-deck-file` | warning | The deck loads a file from another host, which sandboxed and offline players block |
| `no-narration` | warning | Slides without `data-vo` or `aside class="notes"` |
| `long-lesson` | warning | A lesson's narration runs over about 10 minutes |
| `narration-characters` | warning | With `--strict-narration`: narration contains `(`, `)`, `/`, or `→` |
| `unknown-slide` | error | A transcript entry's `id` isn't a slide in the deck |
| `transcript-by-position` | warning | Transcript entries have no `id`, so they're matched by position |
| `transcript-differs` | warning | The transcript's default-language narration differs from the deck |
| `missing-translation` | warning | Slides without narration in one of the course's languages, or no transcript at all |

## Quizzes

| Code | Severity | Meaning |
| --- | --- | --- |
| `duplicate-quiz-id` | error | Two quizzes share an ID |
| `no-questions` | error | The quiz has no questions |
| `out-of-range` | error | `passPercentage` outside 0 to 100, or a negative `maxAttempts` or `points` |
| `missing-id` | error | A question or option has no `id` |
| `duplicate-question-id` | error | Two questions in the quiz share an ID |
| `duplicate-option-id` | error | Two options in a question share an ID |
| `missing-prompt` | error | A question has no prompt |
| `missing-explanation` | error | A question has no explanation |
| `missing-text` | error | An option has no text |
| `too-few-options` | error | A multiple-choice question has fewer than two options |
| `no-correct-option` | error | No option is marked correct |
| `missing-answer` | error | A true/false question has no `correctAnswer`, or a short-answer question no `acceptedAnswers` |
| `wrong-answer-field` | error | A question carries another type's answer field |
| `answer-pattern` | warning | Without `shuffleOptions`, the correct option is always in the same position |

## Published state

| Code | Severity | Meaning |
| --- | --- | --- |
| `published-id-removed` | error | A lesson or quiz ID in `published.json`, or in a platform's IDs, no longer exists; the message names the likely rename |
