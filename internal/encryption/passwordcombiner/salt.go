package passwordcombiner

import (
	"crypto/rand"
	"io"
)

func randomSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt[:])
	if err != nil {
		return salt, err
	}
	return salt, nil
}
