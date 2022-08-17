// Package noopencryptor looks like an encryptor but doesn't do any encryption
package noopencryptor

func New() NoopEncryptor {
	return NoopEncryptor{}
}

type NoopEncryptor struct{}

func (d NoopEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (d NoopEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
