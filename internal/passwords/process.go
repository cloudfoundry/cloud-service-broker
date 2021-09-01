package passwords

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords/parser"
	"github.com/jinzhu/gorm"
)

type Passwords struct {
	Primary     Password
	Secondaries []Password
}

func ProcessPasswords(input string, encryptionEnabled bool, db *gorm.DB) (Passwords, error) {
	parsedPasswords, err := parse(input, encryptionEnabled)
	if err != nil {
		return Passwords{}, err
	}

	var result Passwords
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
	}

	return result, nil
}

func consolidate(parsed parser.PasswordEntry, db *gorm.DB) (Password, error) {
	loaded, ok, err := findPasswordMetadata(db, parsed.Label)
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

	if loaded.Primary != parsed.Primary {
		return Password{}, fmt.Errorf("password migration not implemented yet")
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
		Primary: parsed.Primary,
	})
	if err != nil {
		return Password{}, err
	}

	return result, nil
}
