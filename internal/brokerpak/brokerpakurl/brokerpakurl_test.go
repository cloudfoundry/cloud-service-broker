package brokerpakurl_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/brokerpakurl"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"
)

func TestTerraformResource_URL(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Unable to get current working dir %v", err)
	}
	cases := map[string]struct {
		Resource    manifest.TerraformResource
		Plat        platform.Platform
		ExpectedURL string
	}{
		"default": {
			Resource: manifest.TerraformResource{
				Name:    "foo",
				Version: "1.0",
				Source:  "github.com/myproject",
			},
			Plat: platform.Platform{
				Os:   "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip", "foo", "1.0", "foo", "1.0", "my_os", "my_arch"),
		},
		"custom": {
			Resource: manifest.TerraformResource{
				Name:        "foo",
				Version:     "1.0",
				Source:      "github.com/myproject",
				URLTemplate: "https://myproject/${name}_${version}_${os}_${arch}",
			},
			Plat: platform.Platform{
				Os:   "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("https://myproject/%s_%s_%s_%s", "foo", "1.0", "my_os", "my_arch"),
		},
		"handles_relative_path": {
			Resource: manifest.TerraformResource{
				Name:        "foo",
				Version:     "1.0",
				Source:      "github.com/myproject",
				URLTemplate: "../test_path",
			},
			Plat: platform.Platform{
				Os:   "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("%s/test_path", filepath.Dir(wd)),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			u := brokerpakurl.URL(tc.Resource, tc.Plat)
			if u != tc.ExpectedURL {
				t.Errorf("Expected URL to be %v, got %v", tc.ExpectedURL, u)
			}
		})
	}
}
