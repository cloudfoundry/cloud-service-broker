package stableuuid

import (
	"crypto/sha256"

	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

// FromStrings generates a UUID given two strings. For the same strings, the UUID
// will always be the same.
func FromStrings(s1, s2 string) string {
	var data [16]byte
	copy(data[:], pbkdf2.Key([]byte(s1), []byte(s2), 10, 16, sha256.New))
	return uuid.Must(uuid.FromBytes(data[:])).String()
}
