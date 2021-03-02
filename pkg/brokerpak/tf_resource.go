// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package brokerpak

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

// HashicorpUrlTemplate holds the default template for Hashicorp's terraform binary archive downloads.
const HashicorpUrlTemplate = "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip"

// TerraformResource represents a downloadable binary dependency (Terraform
// version or Provider).
type TerraformResource struct {
	// Name holds the name of this resource. e.g. terraform-provider-google-beta
	Name string `yaml:"name"`

	// Version holds the version of the resource e.g. 1.19.0
	Version string `yaml:"version"`

	// Source holds the URI of an archive that contains the source code for this release.
	Source string `yaml:"source"`

	// UrlTemplate holds a custom URL template to get the release of the given tool.
	// Parameters available are ${name}, ${version}, ${os}, and ${arch}.
	// If non is specified HashicorpUrlTemplate is used.
	UrlTemplate string `yaml:"url_template,omitempty"`
}

var _ validation.Validatable = (*TerraformResource)(nil)

// Validate implements validation.Validatable.
func (tr *TerraformResource) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(tr.Name, "name"),
		validation.ErrIfBlank(tr.Version, "version"),
		validation.ErrIfBlank(tr.Source, "source"),
	)
}

func isURL(path string) bool {
	_, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}

	u, err := url.Parse(path)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// Url constructs a download URL based on a platform.
func (tr *TerraformResource) Url(platform Platform) string {
	replacer := strings.NewReplacer("${name}", tr.Name, "${version}", tr.Version, "${os}", platform.Os, "${arch}", platform.Arch)
	var url string

	if tr.UrlTemplate == "" {
		url = HashicorpUrlTemplate
	} else if isURL(tr.UrlTemplate) {
		url = tr.UrlTemplate
	} else {
		url, _ = filepath.Abs(tr.UrlTemplate)
	}

	return replacer.Replace(url)
}
