package compoundencryptor

func New(primary Encryptor, secondaries ...Encryptor) Encryptor {
	return CompoundEncryptor{
		encryptor:  primary,
		decryptors: secondaries,
	}
}

type CompoundEncryptor struct {
	encryptor  Encryptor
	decryptors []Encryptor
}

func (c CompoundEncryptor) Encrypt(plaintext []byte) (string, error) {
	return c.encryptor.Encrypt(plaintext)
}

func (c CompoundEncryptor) Decrypt(ciphertext string) (data []byte, err error) {
	for _, d := range c.decryptors {
		data, err = d.Decrypt(ciphertext)
		if err == nil {
			return data, nil
		}
	}

	return nil, err
}
