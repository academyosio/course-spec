# Decks and Narration

A deck is one HTML page per section holding its slides and their narration.
course-spec follows [videoascode](https://github.com/grokify/videoascode)'s
conventions, so the same deck is validated by `coursespec`, recorded to video
by `vac`, and played by platforms such as AcademyOS.

## Slides

```html
<section class="slide" data-id="title" data-title="Designing Endpoints"
  data-vo="Welcome to designing endpoints. [PAUSE:500]">
  <h1>Designing Endpoints</h1>
</section>
<section class="slide" data-id="resources" data-lesson="Resources"
  data-vo="An API is a set of resources.">
  <h2>Resources</h2>
</section>
<section class="slide" data-id="methods" data-lesson="Methods">
  <h2>Methods</h2>
  <aside class="notes">GET reads a resource. PUT replaces it.</aside>
</section>
```

- **A slide** is an element with class `slide` and a `data-id`, unique in the
  deck. Slides nested inside another slide aren't counted.
- **Narration** in the course's default language is the slide's `data-vo`,
  or its `aside class="notes"` element. The deck is the only source for that
  language.
- **`[PAUSE:ms]`** directives are timing for the recording tool, not spoken
  text.
- **The deck opens at a slide** named by the URL fragment, such as
  `#methods`, and moves when the fragment changes.

## Lessons

A slide with `data-lesson` starts a lesson. Its value is the lesson's title;
when empty, the slide's `data-title` or ID is used. The lesson's ID is the
slide's `data-id`, so it must be lowercase kebab-case and must never change
once published: learner progress and published videos are keyed by it.

Slides before the first lesson start belong to the first lesson, so a title
slide joins it. A deck with no lesson starts is one lesson with the section's
ID.

In the example above, `title` and `resources` make the lesson `resources`,
and `methods` is a lesson of its own.

## Self-contained decks

Every file a deck loads, styles, scripts, fonts, images, and media, is in the
course directory or its theme and loaded by a relative path. Players that
show decks in a sandboxed frame, and offline playback, block network
requests, so a stylesheet from a CDN or a font from a font service fails
there.

- A relative file that doesn't exist is an error.
- A file loaded from another host is a warning, with the fix to download it
  into the course or theme.
- Links a learner clicks (`<a href>`) aren't loads and may point anywhere.

## Translations

Narration in other languages lives next to the deck in
`deck/voiceover/transcript.json`, in the videoascode transcript format:

```json
{
  "version": "1.0",
  "metadata": {"title": "Designing Endpoints", "defaultLanguage": "en-US", "defaultVoice": {"voiceId": "…"}},
  "slides": [
    {"id": "resources", "index": 1, "transcripts": {
      "es-ES": {"segments": [{"text": "Una API es un conjunto de recursos."}]}}},
    {"id": "methods", "index": 2, "transcripts": {
      "es-ES": {"segments": [{"text": "GET lee un recurso."}, {"text": "PUT lo reemplaza."}]}}}
  ]
}
```

- Each slide entry carries `id`, the slide's `data-id`, and translations are
  matched by it. Matching by position would attach every later translation
  to the wrong slide after a slide is added or removed.
- An older transcript without `id`s is matched by position, with a warning
  to add them.
- A slide `id` that isn't in the deck is an error: it catches a translation
  left behind by a renamed or deleted slide.
- A slide with no narration in one of the course's `languages` is a warning,
  because recording tools record a language only when every slide has it.
- Default-language entries, when present, should match the deck; the deck
  wins and the validator warns when they differ.
- `metadata`, such as the voice, is a recording setting that course-spec
  doesn't read.

## Pronunciations

`pronunciations.json` maps a written term to how narration says it, per
language:

```json
{
  "ProductBuildersHQ": {"en-US": "Product Builders H Q"},
  "OAuth": {"en-US": "oh auth"},
  "API": {"en-US": "A P I"}
}
```

Terms match case-insensitively at word boundaries; where terms overlap, the
one with more words, then the longer one, wins. Only what is spoken changes:
captions and transcripts keep the written term.

## Narration warnings

- A slide without narration.
- A lesson whose narration runs over about 10 minutes, at 150 words per
  minute: split it with `data-lesson`.
- With `--strict-narration`, characters text-to-speech reads badly: `(`,
  `)`, `/`, and `→`. Write them out in words.
