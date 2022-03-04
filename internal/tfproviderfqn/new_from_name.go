package tfproviderfqn

import (
	"fmt"
	"strings"
)

func newFromName(name string) (TfProviderFQN, error) {
	if !strings.HasPrefix(name, prefix) {
		return TfProviderFQN{}, fmt.Errorf("name must have prefix: %s", prefix)
	}

	return TfProviderFQN{
		Hostname:  defaultRegistry,
		Namespace: defaultNamespace,
		Type:      name[len(prefix):],
	}, nil
}
