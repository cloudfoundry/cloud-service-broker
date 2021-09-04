package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
)

// CanaryInput is the value that is encrypted with the key and stored in the database
// to check that the key has not changed. Because we encrypt with a nonce, it's not
// possible to create a rainbow table for this.
const CanaryInput = "canary value"

func encryptCanary(key [32]byte) (string, error) {
	return encryption.NewGCMEncryptor(key).Encrypt([]byte(CanaryInput))
}

func decryptCanary(key [32]byte, canary, label string) error {
	_, err := encryption.NewGCMEncryptor(key).Decrypt(canary)
	switch {
	case err == nil:
		return nil
	case err.Error() == "cipher: message authentication failed": // Unfortunately type is errors.errorString so we cannot do a safer type check
		return fmt.Errorf("canary mismatch for password labeled %q - check that the password value has not changed", label)
	default:
		return err
	}
}
