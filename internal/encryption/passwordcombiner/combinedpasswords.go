package passwordcombiner

import "github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption/gcmencryptor"

type CombinedPassword struct {
	Label             string
	Secret            string
	Salt              []byte
	Encryptor         gcmencryptor.GCMEncryptor
	configuredPrimary bool
	storedPrimary     bool
}

type CombinedPasswords []CombinedPassword

func (c CombinedPasswords) ConfiguredPrimary() (CombinedPassword, bool) {
	for _, p := range c {
		if p.configuredPrimary {
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
