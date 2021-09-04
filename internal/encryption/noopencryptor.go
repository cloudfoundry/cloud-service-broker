package encryption

func NewNoopEncryptor() NoopEncryptor {
	return NoopEncryptor{}
}

type NoopEncryptor struct{}

func (d NoopEncryptor) Encrypt(plaintext []byte) (string, error) {
	return string(plaintext), nil
}

func (d NoopEncryptor) Decrypt(ciphertext string) ([]byte, error) {
	return []byte(ciphertext), nil
}
