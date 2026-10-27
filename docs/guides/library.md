# Using the Go Library

```bash
go get github.com/academyosio/course-spec@v0.1.0
```

| Package | Purpose |
| --- | --- |
| `github.com/academyosio/course-spec` (package `coursespec`) | Types for every file, `Load`, `Find`, publish status, and published-state reading and writing |
| `github.com/academyosio/course-spec/deck` | Slides, lessons, transcripts, and content fingerprints from HTML decks |
| `github.com/academyosio/course-spec/validate` | Every check, as findings with a fix |
| `github.com/academyosio/course-spec/schema` | The generated JSON Schemas, embedded |

The module's dependencies are small, so platforms can import it: an HTML
parser for decks, and the schema generator only at generation time.

## Validate

```go
import "github.com/academyosio/course-spec/validate"

findings := validate.Course("courses/building-web-apis", validate.Options{})
errs, warns := validate.Counts(findings)
for _, f := range findings {
    fmt.Println(f.String())
}
```

Platforms import the same function, so a course checked locally passes the
same checks on import.

## Load

```go
import coursespec "github.com/academyosio/course-spec"

c, err := coursespec.Load("courses/building-web-apis")
if err != nil {
    return err // the first problem; use validate for all of them
}
for _, s := range c.Sections {
    fmt.Println(s.ID(), s.Title(), s.Ref.IsReleased())
    for _, l := range s.Lessons {
        fmt.Println("  ", l.ID, l.Type, len(l.Slides), c.Fingerprint(s, l))
    }
    if s.Quiz != nil {
        fmt.Println("  quiz", s.Quiz.ID, coursespec.QuizFingerprint(s.Quiz))
    }
}
```

`Load` resolves lessons from decks with the manifests' details applied,
matches translations to slides, and reads quizzes, pronunciations, and
published state. Unreleased sections without files load with only their ID
and title.

## Find courses

```go
dirs, err := coursespec.Find("courses") // every directory holding a course.json
```

## Publish status

```go
report := c.Status() // sections, lessons, and quizzes as new, changed, published, retired, or planned
```

## Decks on their own

```go
import "github.com/academyosio/course-spec/deck"

d, err := deck.ParseFile("deck/index.html")
for _, l := range d.Lessons("section-id") {
    fmt.Println(l.ID, l.Title, len(l.Slides), l.Words())
}
for _, r := range d.Refs {
    if r.Remote() {
        fmt.Println("loads from another host:", r.URL)
    }
}
```

## Schemas

```go
import "github.com/academyosio/course-spec/schema"

data, err := schema.Get("quiz")
```
