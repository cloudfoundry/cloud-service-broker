package compoundencryptor

func New(encryptor Encryptor, decryptors ...Encryptor) Encryptor {
	return CompoundEncryptor{
		encryptor:  encryptor,
		decryptors: decryptors,
	}
}

type CompoundEncryptor struct {
	encryptor  Encryptor
	decryptors []Encryptor
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
