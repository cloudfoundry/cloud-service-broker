// Package brokerpakurl handles the logic of working out which URL
// to fetch Terraform resources from
package brokerpakurl

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
)

const (
	HashicorpURLTemplate = "https://releases.hashicorp.com/${name}/${version}/${name}_${version}_${os}_${arch}.zip"
	TofuURLTemplate      = "https://github.com/opentofu/opentofu/releases/download/v${version}/tofu_${version}_${os}_${arch}.zip"
)

func HashicorpURL(name, version, urlTemplate string, plat platform.Platform) string {
	return replaceURL(name, version, urlTemplate, HashicorpURLTemplate, plat)
}

func TofuURL(name, version, urlTemplate string, plat platform.Platform) string {
	return replaceURL(name, version, urlTemplate, TofuURLTemplate, plat)
}

func replaceURL(name, version, urlTemplate, defaultTemplate string, plat platform.Platform) string {
	replacer := strings.NewReplacer("${name}", name, "${version}", version, "${os}", plat.Os, "${arch}", plat.Arch)
	var template string

	switch {
	case urlTemplate == "":
		template = defaultTemplate
	case isURL(urlTemplate):
		template = urlTemplate
	default:
		template, _ = filepath.Abs(urlTemplate)
	}

	return replacer.Replace(template)
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
