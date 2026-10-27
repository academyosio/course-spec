# Published State and Fingerprints

Source, build output, and published state are kept apart:

- **Source** is what authors write: manifests, decks, transcripts, quizzes.
- **Build output** (clips, videos, captions) is generated into each
  section's `build/` and not committed.
- **Published state** is `published.json` in the course directory, written by
  tools after recording, uploading, or importing, and committed, so a fresh
  clone knows what is live. Authors don't edit it.

## `published.json`

```json
{
  "courseVersion": "1.2.0",
  "lessons": {
    "resources": {
      "contentFingerprint": "sha256:…",
      "recordingFingerprint": "3f9a1c0b7d2e4a65",
      "durationMs": 151000,
      "videos": {"en-US": {"provider": "youtube", "id": "dQw4w9WgXcQ"}},
      "captions": {"en-US": {"provider": "youtube"}}
    }
  },
  "quizzes": {
    "designing-endpoints-quiz": {"fingerprint": "sha256:…"}
  },
  "video": {
    "playlists": {"en-US": {"provider": "youtube", "id": "PL…"}}
  },
  "platforms": {
    "academyos": {
      "baseUrl": "https://app.academyos.io",
      "courseId": "…",
      "lessons": {"resources": "…"},
      "quizzes": {"designing-endpoints-quiz": "…"},
      "revision": 4
    }
  }
}
```

- Keys are course-spec IDs; platform IDs are values, so one course can be
  published to several platforms.
- A platform's `revision` is its revision of the course when this copy was
  last synced, so the platform can refuse a push from a stale copy.
- `recordingFingerprint` belongs to the recording tool, in whatever form it
  uses; course-spec doesn't interpret it.
- No secrets, tokens, or private URLs belong in this file.

## Content fingerprints

A lesson's content fingerprint, written `sha256:<hex>`, covers everything an
author writes and nothing a recording tool chooses. For each of the lesson's
slides, in order, it hashes:

1. the slide's HTML, without the slide element's own `data-id`;
2. for each language in the course's `languages`, sorted by code, the slide's
   narration: from the deck for the default language, from the transcript
   for the others, with segments joined by single spaces.

Narration is normalized first: `[PAUSE:ms]` directives are removed and runs
of whitespace become one space. Each value is written with its length before
it, so boundaries can't shift.

What follows from that:

- **Splitting the same sentences differently** isn't a change.
- **Renaming a slide** isn't a change, so a renamed lesson keeps its
  fingerprint, and the validator can say which new ID it was likely renamed
  to.
- **The voice, pronunciations, timing, and resolution** aren't part of it;
  they belong in the recording tool's own fingerprint.
- **Theme files aren't part of it,** so a theme change marks no lesson as
  changed. Re-record affected lessons after changing a theme.

A quiz's fingerprint hashes its settings and questions.

## What the fingerprints tell you

`coursespec lessons` compares each lesson and quiz with `published.json`:

| Status | Meaning |
| --- | --- |
| `new` | Never published |
| `changed` | Its content differs from what was published |
| `published` | Unchanged since it was published |
| `retired` | Kept for its published ID |
| `planned` | In an unreleased section |

`coursespec validate` uses them too:

- A published ID that no longer exists is an error, naming the likely rename
  when a new lesson has the old fingerprint. Restore the ID, or keep the
  lesson with `"retired": true`.
- Content that changed while `version` still equals the published
  `courseVersion` is a warning to raise the version.

## Writing it from a tool

Go tools write the file with `coursespec.WritePublished`, which writes the
indented form that's committed:

```go
pub, err := coursespec.ReadPublished("course/published.json")
// update pub.Platforms["academyos"], pub.Lessons, …
err = coursespec.WritePublished("course/published.json", pub)
```
