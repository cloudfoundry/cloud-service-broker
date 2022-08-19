package passwordcombiner

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/gcmencryptor"
)

// CanaryInput is the value that is encrypted with the key and stored in the database
// to check that the key has not changed. Because we encrypt with a nonce, it's not
// possible to create a rainbow table for this.
const CanaryInput = "canary value"

func encryptCanary(encryptor gcmencryptor.GCMEncryptor) ([]byte, error) {
	return encryptor.Encrypt([]byte(CanaryInput))
}

func decryptCanary(encryptor gcmencryptor.GCMEncryptor, canary []byte, label string) error {
	_, err := encryptor.Decrypt(canary)
	switch {
	case err == nil:
		return nil
	// Unfortunately type is errors.errorString, so we cannot do a safer type check
	case err.Error() == "cipher: message authentication failed", err.Error() == "malformed ciphertext":
		return fmt.Errorf("canary mismatch for password labeled %q - check that the password value has not changed", label)
	default:
		return fmt.Errorf("encryption canary error: %w", err)
	}
}
