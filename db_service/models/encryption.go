package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	// Initialize an AES block cipher
	block, err := aes.NewCipher(key[:])
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

func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	// Initialize an AES block cipher
	block, err := aes.NewCipher(key[:])
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
