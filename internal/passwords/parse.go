package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordparser"
)

func parse(input string, encryptionEnabled bool) ([]passwordparser.PasswordEntry, error) {
	parsedPasswords, err := passwordparser.Parse(input)
	count := primaries(parsedPasswords)
	switch {
	case err != nil:
		return nil, err
	case encryptionEnabled && count == 0:
		return nil, fmt.Errorf("encryption is enabled but no primary password is set")
	case !encryptionEnabled && count != 0:
		return nil, fmt.Errorf("encryption is disabled but a primary password is set")
	default:
		return parsedPasswords, nil
	}
}

func primaries(passwords []passwordparser.PasswordEntry) (count int) {
	for _, p := range passwords {
		if p.Primary {
			count++
		}
	}
	return count
}
