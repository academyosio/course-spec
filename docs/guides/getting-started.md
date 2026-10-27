# Getting Started

## Install

`coursespec` is a Go command. With Go 1.26 or later:

```bash
go install github.com/academyosio/course-spec/cmd/coursespec@latest
coursespec --help
```

## Check the example course

The repository includes a complete small course,
[`testdata/courses/web-apis`](https://github.com/academyosio/course-spec/tree/main/testdata/courses/web-apis):
a deck with two lessons and a text lesson, a Spanish translation, a quiz, a
quiz-only final assessment, and a planned section.

```bash
git clone https://github.com/academyosio/course-spec
cd course-spec
coursespec validate testdata/courses/web-apis
coursespec lessons testdata/courses/web-apis
```

```text
testdata/courses/web-apis: 0 errors, 0 warnings

building-web-apis 1.0.0 (Building Web APIs)
  designing-endpoints: Designing Endpoints
    new        resources: Resources (2 slides)
    new        glossary: Glossary (text)
    new        methods: Methods (1 slides)
    new        designing-endpoints-quiz: Designing Endpoints Quiz (quiz, 3 questions)
  final-assessment: Final Assessment
    new        final-assessment-quiz: Final Assessment (quiz, 1 questions)
  versioning: Versioning  [planned]
```

Every lesson is `new` because nothing has been published yet.

## Start your own course

1. Make a directory with a `course.json`:

    ```json
    {
      "specVersion": "1",
      "id": "my-first-course",
      "version": "0.1.0",
      "title": "My First Course",
      "sections": [{"path": "sections/01-introduction"}]
    }
    ```

2. Add the section's `sections/01-introduction/section.json`:

    ```json
    {
      "id": "introduction",
      "title": "Introduction",
      "source": {"type": "htmlDeck", "path": "deck/index.html"}
    }
    ```

3. Write `sections/01-introduction/deck/index.html` with slides, each a
   `class="slide"` element with a `data-id` and narration in `data-vo`. A
   slide with `data-lesson` starts a lesson:

    ```html
    <section class="slide" data-id="welcome" data-lesson="Welcome"
      data-vo="Welcome to my first course.">
      <h1>Welcome</h1>
    </section>
    ```

4. Run `coursespec validate .` and fix what it reports.

Next, read [The course directory](course-directory.md) for every field, or
[Decks and narration](decks-and-narration.md) for how lessons come from
slides.
