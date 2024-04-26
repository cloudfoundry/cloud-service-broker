// Package compoundencryptor allows encryptors to be combined
package compoundencryptor

import "github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"

func New(encryptor storage.Encryptor, decryptors ...storage.Encryptor) storage.Encryptor {
	return CompoundEncryptor{
		encryptor:  encryptor,
		decryptors: decryptors,
	}
}

type CompoundEncryptor struct {
	encryptor  storage.Encryptor
	decryptors []storage.Encryptor
}

func (c CompoundEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	return c.encryptor.Encrypt(plaintext)
}

func (c CompoundEncryptor) Decrypt(ciphertext []byte) (data []byte, err error) {
	for _, d := range c.decryptors {
		data, err = d.Decrypt(ciphertext)
		if err == nil {
			return data, nil
		}
	}

	return nil, err
}
