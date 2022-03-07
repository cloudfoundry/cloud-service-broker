package tfproviderfqn

import (
	"fmt"
	"strings"
)

func newFromProvider(provider string) (TfProviderFQN, error) {
	parts := strings.Split(provider, "/")
	switch len(parts) {
	case 1:
		return TfProviderFQN{
			Hostname:  defaultRegistry,
			Namespace: defaultNamespace,
			Type:      parts[0],
		}, nil
	case 2:
		return TfProviderFQN{
			Hostname:  defaultRegistry,
			Namespace: parts[0],
			Type:      parts[1],
		}, nil
	case 3:
		return TfProviderFQN{
			Hostname:  parts[0],
			Namespace: parts[1],
			Type:      parts[2],
		}, nil
	default:
		return TfProviderFQN{}, fmt.Errorf("invalid format; valid format is [<HOSTNAME>/]<NAMESPACE>/<TYPE>")
	}
}
