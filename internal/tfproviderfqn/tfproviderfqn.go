// Package tfproviderfqn implements fully qualified Terraform provider names
package tfproviderfqn

import "fmt"

const (
	prefix           = "terraform-provider-"
	defaultRegistry  = "registry.opentofu.org"
	defaultNamespace = "hashicorp"
)

func New(name, provider string) (TfProviderFQN, error) {
	switch provider {
	case "":
		return newFromName(name)
	default:
		return newFromProvider(provider)
	}
}

func Must(name, provider string) TfProviderFQN {
	fqn, err := New(name, provider)
	switch err {
	case nil:
		return fqn
	default:
		panic(err)
	}
}

type TfProviderFQN struct {
	Hostname  string
	Namespace string
	Type      string
}

func (t TfProviderFQN) String() string {
	if t.Hostname == "" && t.Namespace == "" && t.Type == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", t.Hostname, t.Namespace, t.Type)
}
