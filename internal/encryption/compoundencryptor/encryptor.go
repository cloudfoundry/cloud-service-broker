package compoundencryptor

type Encryptor interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}
