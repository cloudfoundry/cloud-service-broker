package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type GCMEncryptor struct {
	Key *[32]byte
}

func NewGCMEncryptor(key *[32]byte) GCMEncryptor {
	return GCMEncryptor{Key: key}
}

func (d GCMEncryptor) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	// Initialize an AES block cipher
	block, err := aes.NewCipher(d.Key[:])
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

	// Return the encrypted text appended to the nonce
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (d GCMEncryptor) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	// Initialize an AES block cipher
	block, err := aes.NewCipher(d.Key[:])
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

type NoopEncryptor struct {
}

func (d NoopEncryptor) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return plaintext, nil
}

func (d NoopEncryptor) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	return ciphertext, nil
}

func NewNoopEncryptor() NoopEncryptor {
	return NoopEncryptor{}
}
