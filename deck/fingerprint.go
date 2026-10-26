package deck

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"slices"
	"strconv"
)

// FingerprintPrefix starts every course-spec fingerprint.
const FingerprintPrefix = "sha256:"

// Fingerprint returns the lesson's content fingerprint. For each slide in
// order it hashes the slide's HTML, then for each language sorted by code
// the slide's narration: the deck's for defaultLang and translations' for
// the others. Every value is written after its length, so boundaries can't
// shift. Recording settings such as the voice are not included.
func (l Lesson) Fingerprint(languages []string, defaultLang string, translations Narration) string {
	langs := slices.Clone(languages)
	slices.Sort(langs)
	h := sha256.New()
	for _, s := range l.Slides {
		writeField(h, s.HTML)
		for _, lang := range langs {
			text := s.Narration
			if lang != defaultLang {
				text = translations[s.ID][lang]
			}
			writeField(h, []byte(lang))
			writeField(h, []byte(text))
		}
	}
	return FingerprintPrefix + hex.EncodeToString(h.Sum(nil))
}

func writeField(h hash.Hash, b []byte) {
	// hash.Hash.Write never returns an error.
	_, _ = h.Write([]byte(strconv.Itoa(len(b)) + ":"))
	_, _ = h.Write(b)
}
