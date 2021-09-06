package passwordcombiner

import "github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"

type CombinedPassword struct {
	Label         string
	Secret        string
	Salt          []byte
	Encryptor     gcmencryptor.GCMEncryptor
	parsedPrimary bool
	storedPrimary bool
}

type CombinedPasswords []CombinedPassword

func (c CombinedPasswords) ParsedPrimary() (CombinedPassword, bool) {
	for _, p := range c {
		if p.parsedPrimary {
			return p, true
		}
	}

	return CombinedPassword{}, false
}

func (c CombinedPasswords) StoredPrimary() (CombinedPassword, bool) {
	for _, p := range c {
		if p.storedPrimary {
			return p, true
		}
	}

	return CombinedPassword{}, false
}
