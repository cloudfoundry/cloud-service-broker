package passwords

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

type Password struct {
	Label  string
	Secret string
	Salt   [32]byte
}

func (p Password) Key() (key [32]byte) {
	copy(key[:], pbkdf2.Key([]byte(p.Secret), p.Salt[:], 100000, 32, sha256.New))
	return
}
