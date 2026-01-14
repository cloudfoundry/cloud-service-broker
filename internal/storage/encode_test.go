package storage

import (
	"errors"
	"testing"
)

// mockEncryptor is a simple mock for testing encode/decode functions directly
type mockEncryptor struct {
	decryptFunc      func([]byte) ([]byte, error)
	encryptFunc      func([]byte) ([]byte, error)
	decryptCallCount int
	encryptCallCount int
	lastDecryptArg   []byte
	lastEncryptArg   []byte
}

func (m *mockEncryptor) Decrypt(data []byte) ([]byte, error) {
	m.decryptCallCount++
	m.lastDecryptArg = data
	if m.decryptFunc != nil {
		return m.decryptFunc(data)
	}
	return data, nil
}

func (m *mockEncryptor) Encrypt(data []byte) ([]byte, error) {
	m.encryptCallCount++
	m.lastEncryptArg = data
	if m.encryptFunc != nil {
		return m.encryptFunc(data)
	}
	return data, nil
}

// TestDecodeBytesHandlesEmptyInput verifies that decodeBytes gracefully handles
// nil and empty byte slices without calling the encryptor. This is critical for
// handling database records migrated from V1 schema where bind_resource column
// was added but is NULL for existing records (see PR #1341).
func TestDecodeBytesHandlesEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "nil input", input: nil},
		{name: "empty slice", input: []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor := &mockEncryptor{}
			store := New(nil, encryptor)

			result, err := store.decodeBytes(tt.input)
			if err != nil {
				t.Errorf("decodeBytes(%v) returned error: %v", tt.input, err)
			}
			if result != nil {
				t.Errorf("decodeBytes(%v) = %v, want nil", tt.input, result)
			}
			if encryptor.decryptCallCount != 0 {
				t.Errorf("decodeBytes(%v) called Decrypt %d times, but should not have", tt.input, encryptor.decryptCallCount)
			}
		})
	}
}

// TestEncodeBytesHandlesEmptyInput verifies that encodeBytes gracefully handles
// nil and empty byte slices without calling the encryptor.
func TestEncodeBytesHandlesEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "nil input", input: nil},
		{name: "empty slice", input: []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor := &mockEncryptor{}
			store := New(nil, encryptor)

			result, err := store.encodeBytes(tt.input)
			if err != nil {
				t.Errorf("encodeBytes(%v) returned error: %v", tt.input, err)
			}
			if result != nil {
				t.Errorf("encodeBytes(%v) = %v, want nil", tt.input, result)
			}
			if encryptor.encryptCallCount != 0 {
				t.Errorf("encodeBytes(%v) called Encrypt %d times, but should not have", tt.input, encryptor.encryptCallCount)
			}
		})
	}
}

// TestDecodeBytesCallsDecryptorForNonEmptyInput verifies that decodeBytes
// correctly calls the encryptor for non-empty input.
func TestDecodeBytesCallsDecryptorForNonEmptyInput(t *testing.T) {
	encryptor := &mockEncryptor{
		decryptFunc: func(data []byte) ([]byte, error) {
			return []byte("decrypted"), nil
		},
	}
	store := New(nil, encryptor)

	input := []byte("encrypted-data")
	result, err := store.decodeBytes(input)

	if err != nil {
		t.Errorf("decodeBytes returned error: %v", err)
	}
	if string(result) != "decrypted" {
		t.Errorf("decodeBytes = %s, want decrypted", result)
	}
	if encryptor.decryptCallCount != 1 {
		t.Errorf("Decrypt called %d times, want 1", encryptor.decryptCallCount)
	}
	if string(encryptor.lastDecryptArg) != string(input) {
		t.Errorf("Decrypt called with %v, want %v", encryptor.lastDecryptArg, input)
	}
}

// TestEncodeBytesCallsEncryptorForNonEmptyInput verifies that encodeBytes
// correctly calls the encryptor for non-empty input.
func TestEncodeBytesCallsEncryptorForNonEmptyInput(t *testing.T) {
	encryptor := &mockEncryptor{
		encryptFunc: func(data []byte) ([]byte, error) {
			return []byte("encrypted"), nil
		},
	}
	store := New(nil, encryptor)

	input := []byte("plaintext-data")
	result, err := store.encodeBytes(input)

	if err != nil {
		t.Errorf("encodeBytes returned error: %v", err)
	}
	if string(result) != "encrypted" {
		t.Errorf("encodeBytes = %s, want encrypted", result)
	}
	if encryptor.encryptCallCount != 1 {
		t.Errorf("Encrypt called %d times, want 1", encryptor.encryptCallCount)
	}
	if string(encryptor.lastEncryptArg) != string(input) {
		t.Errorf("Encrypt called with %v, want %v", encryptor.lastEncryptArg, input)
	}
}

// TestDecodeBytesReturnsDecryptionError verifies that decodeBytes properly
// wraps and returns decryption errors.
func TestDecodeBytesReturnsDecryptionError(t *testing.T) {
	encryptor := &mockEncryptor{
		decryptFunc: func(data []byte) ([]byte, error) {
			return nil, errors.New("malformed ciphertext")
		},
	}
	store := New(nil, encryptor)

	_, err := store.decodeBytes([]byte("bad-data"))

	if err == nil {
		t.Error("decodeBytes should return error for decryption failure")
	}
	expectedMsg := "decryption error: malformed ciphertext"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

// TestEncodeBytesReturnsEncryptionError verifies that encodeBytes properly
// wraps and returns encryption errors.
func TestEncodeBytesReturnsEncryptionError(t *testing.T) {
	encryptor := &mockEncryptor{
		encryptFunc: func(data []byte) ([]byte, error) {
			return nil, errors.New("encryption failed")
		},
	}
	store := New(nil, encryptor)

	_, err := store.encodeBytes([]byte("data"))

	if err == nil {
		t.Error("encodeBytes should return error for encryption failure")
	}
	expectedMsg := "encryption error: encryption failed"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}
