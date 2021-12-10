package storage

import "encoding/json"

func (s *Storage) marshalAndEncrypt(a interface{}) ([]byte, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	c, err := s.encryptor.Encrypt(b)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Storage) decryptAndUnmarshalObject(a []byte) (map[string]interface{}, error) {
	b, err := s.encryptor.Decrypt(a)
	if err != nil {
		return nil, err
	}

	var receiver map[string]interface{}
	if err := json.Unmarshal(b, &receiver); err != nil {
		return nil, err
	}

	return receiver, nil
}
