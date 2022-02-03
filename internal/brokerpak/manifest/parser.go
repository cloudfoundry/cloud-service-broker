package manifest

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func Parse(input []byte) (*Manifest, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	decoder.KnownFields(true)

	var receiver Manifest
	if err := decoder.Decode(&receiver); err != nil {
		return nil, fmt.Errorf("error parsing manifest: %w", err)
	}

	if err := receiver.Validate(); err != nil {
		return nil, fmt.Errorf("error validating manifest: %w", err)
	}

	return &receiver, nil
}
