package storage

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . Encryptor
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}
