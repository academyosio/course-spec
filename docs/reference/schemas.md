# JSON Schemas

Each course-spec file has a JSON Schema (draft 2020-12), generated from the
Go types, which are the source of truth. Unknown properties are not allowed.

| File | Schema | `$id` |
| --- | --- | --- |
| `course.json` | `schema/course.schema.json` | `https://academyos.io/schemas/course-spec/v1/course.schema.json` |
| `section.json` | `schema/section.schema.json` | `https://academyos.io/schemas/course-spec/v1/section.schema.json` |
| `quiz.json` | `schema/quiz.schema.json` | `https://academyos.io/schemas/course-spec/v1/quiz.schema.json` |
| `published.json` | `schema/published.schema.json` | `https://academyos.io/schemas/course-spec/v1/published.schema.json` |

Print one with `coursespec schema <name>`, or read it from Go with
`schema.Get(name)`.

The schemas check each file's shape. Rules that span files, such as a
lesson named in `section.json` existing in the deck, or a question's
`lesson` existing in the course, are checked by `coursespec validate`, which
also applies every schema rule.

`pronunciations.json` is a plain object of objects, term to language to
spoken form, and has no schema of its own. Transcripts use the videoascode
transcript format, defined by its schema in that project.

## Regenerating

After changing the Go types:

```bash
go generate ./schema
```

CI fails when the committed schemas don't match the types.
