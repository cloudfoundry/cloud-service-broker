package gcmencryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func New(key [32]byte) GCMEncryptor {
	return GCMEncryptor{key: key[:]}
}

type GCMEncryptor struct {
	key []byte
}

func (e GCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	e.validate()

	// Initialize an AES block cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// Specify a GCM block cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// Return the encrypted text appended to the nonce, encoded in b64
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return sealed, nil
}

func (e GCMEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	e.validate()

	// Initialize an AES block cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// Specify a GCM block cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	// The encrypted text comes after the nonce
	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

func (e GCMEncryptor) validate() {
	if len(e.key) != 32 {
		panic("encryption method called on uninitialised encryptor")
	}
}
