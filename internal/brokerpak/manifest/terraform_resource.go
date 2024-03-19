package manifest

import (
	"strings"
)

type TerraformResource struct {
	// Name holds the name of this resource. e.g. terraform-provider-google-beta
	Name string `yaml:"name"`

	// Version holds the version of the resource e.g. 1.19.0
	Version string `yaml:"version"`

	// Source holds the URI of an archive that contains the source code for this release.
	Source string `yaml:"source"`

	// Provider holds path to extract the provider path.
	Provider string `yaml:"provider,omitempty"`

	// URLTemplate holds a custom URL template to get the release of the given tool.
	// Parameters available are ${name}, ${version}, ${os}, and ${arch}.
	// If non is specified HashicorpUrlTemplate is used.
	URLTemplate string `yaml:"url_template,omitempty"`

	// Default is used to mark the default Terraform version when there is more than one
	Default bool `yaml:"default,omitempty"`
}

type terraformResourceType int

const (
	invalidType terraformResourceType = iota
	terraformVersion
	terraformProvider
	otherBinary
)

func (tr *TerraformResource) resourceType() terraformResourceType {
	switch {
	case tr.Name == "":
		return invalidType
	case tr.Name == binaryName:
		return terraformVersion
	case strings.HasPrefix(tr.Name, "terraform-provider-"):
		return terraformProvider
	default:
		return otherBinary
	}
}
