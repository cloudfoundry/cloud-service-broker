package storage

import (
	"encoding/json"
	"fmt"
)

func (s *Storage) encodeBytes(b []byte) ([]byte, error) {
	c, err := s.encryptor.Encrypt(b)
	if err != nil {
		return nil, fmt.Errorf("encryption error: %w", err)
	}

	return c, nil
}

func (s *Storage) encodeJSON(a interface{}) ([]byte, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("JSON marshal error: %w", err)
	}

	return s.encodeBytes(b)
}

func (s *Storage) decodeBytes(a []byte) ([]byte, error) {
	d, err := s.encryptor.Decrypt(a)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %w", err)
	}

	return d, nil
}

func (s *Storage) decodeJSONObject(a []byte) (JSONObject, error) {
	b, err := s.decodeBytes(a)
	switch {
	case err != nil:
		return nil, err
	case len(b) == 0:
		return nil, nil
	}

	var receiver JSONObject
	if err := json.Unmarshal(b, &receiver); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return receiver, nil
}
