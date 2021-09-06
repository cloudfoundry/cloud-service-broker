package gcmencryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"errors"
	"io"
)

func New(key [32]byte) GCMEncryptor {
	return GCMEncryptor{key: key}
}

type GCMEncryptor struct {
	key [32]byte
}

func (d GCMEncryptor) Encrypt(plaintext []byte) (string, error) {
	// Initialize an AES block cipher
	block, err := aes.NewCipher(d.key[:])
	if err != nil {
		return "", err
	}

	// Specify a GCM block cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	// Return the encrypted text appended to the nonce, encoded in b64
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return b64.StdEncoding.EncodeToString(sealed), nil
}

func (d GCMEncryptor) Decrypt(ciphertext string) ([]byte, error) {
	decoded, err := b64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	// Initialize an AES block cipher
	block, err := aes.NewCipher(d.key[:])
	if err != nil {
		return nil, err
	}

	// Specify a GCM block cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(decoded) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	// The encrypted text comes after the nonce
	return gcm.Open(nil,
		decoded[:gcm.NonceSize()],
		decoded[gcm.NonceSize():],
		nil,
	)
}
