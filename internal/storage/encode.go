package storage

import "encoding/json"

func (s *Storage) encodeBytes(b []byte) ([]byte, error) {
	c, err := s.encryptor.Encrypt(b)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Storage) encodeJSON(a interface{}) ([]byte, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return s.encodeBytes(b)
}

func (s *Storage) decodeBytes(a []byte) ([]byte, error) {
	return s.encryptor.Decrypt(a)
}

func (s *Storage) decodeJSONObject(a []byte) (map[string]interface{}, error) {
	b, err := s.decodeBytes(a)
	if err != nil {
		return nil, err
	}

	var receiver map[string]interface{}
	if err := json.Unmarshal(b, &receiver); err != nil {
		return nil, err
	}

	return receiver, nil
}
