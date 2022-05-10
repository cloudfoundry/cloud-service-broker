// Package brokerpakurl handles the logic of working out which URL
// to fetch Terraform resources from
package brokerpakurl

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
)

const HashicorpURLTemplate = "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip"

func URL(name, version, urlTemplate string, plat platform.Platform) string {
	replacer := strings.NewReplacer("${name}", name, "${version}", version, "${os}", plat.Os, "${arch}", plat.Arch)
	var url string

	switch {
	case urlTemplate == "":
		url = HashicorpURLTemplate
	case isURL(urlTemplate):
		url = urlTemplate
	default:
		url, _ = filepath.Abs(urlTemplate)
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
