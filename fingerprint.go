package coursespec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/academyosio/course-spec/deck"
)

// QuizFingerprint hashes a quiz's settings and questions, so a changed quiz
// is noticed the way a changed lesson is.
func QuizFingerprint(q *Quiz) string {
	data, err := json.Marshal(q)
	if err != nil {
		panic(fmt.Sprintf("marshal quiz: %v", err)) // a Quiz always marshals
	}
	sum := sha256.Sum256(data)
	return deck.FingerprintPrefix + hex.EncodeToString(sum[:])
}
