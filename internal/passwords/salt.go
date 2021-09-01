package passwords

import (
	"crypto/rand"
	"io"
)

func randomSalt() ([32]byte, error) {
	var salt [32]byte
	_, err := io.ReadFull(rand.Reader, salt[:])
	if err != nil {
		return salt, err
	}
	return salt, nil
}
