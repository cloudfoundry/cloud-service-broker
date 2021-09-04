package encryption

func NewCompoundEncryptor(primary Encryptor, secondaries ...Encryptor) Encryptor {
	return CompoundEncryptor{
		primary:     primary,
		secondaries: secondaries,
	}
}

type CompoundEncryptor struct {
	primary     Encryptor
	secondaries []Encryptor
}

func (c CompoundEncryptor) Encrypt(plaintext []byte) (string, error) {
	return c.primary.Encrypt(plaintext)
}

func (c CompoundEncryptor) Decrypt(ciphertext string) (data []byte, err error) {
	for _, decryptor := range append([]Encryptor{c.primary}, c.secondaries...) {
		data, err = decryptor.Decrypt(ciphertext)
		if err == nil {
			return data, nil
		}
	}

	return nil, err
}
