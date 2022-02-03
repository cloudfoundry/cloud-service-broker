// Package brokerpakurl handles the logic of working out which URL
// to fetch Terraform resources from
package brokerpakurl

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"
)

const HashicorpUrlTemplate = "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip"

func URL(resource manifest.TerraformResource, plat platform.Platform) string {
	replacer := strings.NewReplacer("${name}", resource.Name, "${version}", resource.Version, "${os}", plat.Os, "${arch}", plat.Arch)
	var url string

	switch {
	case resource.URLTemplate == "":
		url = HashicorpUrlTemplate
	case isURL(resource.URLTemplate):
		url = resource.URLTemplate
	default:
		url, _ = filepath.Abs(resource.URLTemplate)
	}

	return replacer.Replace(url)
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
