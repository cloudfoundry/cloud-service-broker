package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/parser"
	"gorm.io/gorm"
)

type Passwords struct {
	Primary        Password
	Secondaries    []Password
	ChangedPrimary bool
}

func ProcessPasswords(input string, encryptionEnabled bool, db *gorm.DB) (Passwords, error) {
	parsedPasswords, err := parse(input, encryptionEnabled)
	if err != nil {
		return Passwords{}, err
	}

	var result Passwords
	labels := make(map[string]struct{})
	for _, p := range parsedPasswords {
		entry, err := consolidate(p, db)
		if err != nil {
			return Passwords{}, err
		}

		if p.Primary {
			result.Primary = entry
		} else {
			result.Secondaries = append(result.Secondaries, entry)
		}

		labels[p.Label] = struct{}{}
	}

	previousPrimary, found, err := findPasswordMetadataForPrimary(db)
	if err != nil {
		return Passwords{}, err
	}
	if found {
		if _, ok := labels[previousPrimary.Label]; !ok {
			return Passwords{}, fmt.Errorf("the previous primary password labeled %q was not specified", previousPrimary.Label)
		}

		result.ChangedPrimary = previousPrimary.Label != result.Primary.Label
	}

	return result, nil
}

func consolidate(parsed parser.PasswordEntry, db *gorm.DB) (Password, error) {
	loaded, ok, err := findPasswordMetadataForLabel(db, parsed.Label)
	switch {
	case err != nil:
		return Password{}, err
	case ok:
		return checkRecord(loaded, parsed)
	default:
		return newRecord(parsed, db)
	}
}

func checkRecord(loaded passwordMetadata, parsed parser.PasswordEntry) (Password, error) {
	result := Password{
		Label:  parsed.Label,
		Secret: parsed.Secret,
		Salt:   loaded.Salt,
	}

	if err := decryptCanary(result.Key(), loaded.Canary, parsed.Label); err != nil {
		return Password{}, err
	}

	return result, nil
}

func newRecord(parsed parser.PasswordEntry, db *gorm.DB) (Password, error) {
	salt, err := randomSalt()
	if err != nil {
		return Password{}, err
	}

	result := Password{
		Label:  parsed.Label,
		Secret: parsed.Secret,
		Salt:   salt,
	}

	canary, err := encryptCanary(result.Key())
	if err != nil {
		return Password{}, err
	}

	err = savePasswordMetadata(db, passwordMetadata{
		Label:   parsed.Label,
		Salt:    salt,
		Canary:  canary,
		Primary: false,
	})
	if err != nil {
		return Password{}, err
	}

	return result, nil
}
