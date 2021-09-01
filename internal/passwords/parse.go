package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/parser"
)

func parse(input string, encryptionEnabled bool) ([]parser.PasswordEntry, error) {
	parsedPasswords, err := parser.Parse(input)
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

func primaries(passwords []parser.PasswordEntry) (count int) {
	for _, p := range passwords {
		if p.Primary {
			count++
		}
	}
	return count
}
