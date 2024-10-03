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

// Package platform is a utility for handling platform data
package platform

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/validation"
)

// CurrentPlatform returns the platform defined by GOOS and GOARCH.
func CurrentPlatform() Platform {
	return Platform{Os: runtime.GOOS, Arch: runtime.GOARCH}
}

func Parse(s string) Platform {
	if s == "current" {
		return CurrentPlatform()
	}
	parts := strings.SplitN(s, "/", 2)
	switch len(parts) {
	case 2:
		return Platform{
			Os:   parts[0],
			Arch: parts[1],
		}
	default:
		return Platform{}
	}
}

// Platform holds an os/architecture pair.
type Platform struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

var _ validation.Validatable = (*Platform)(nil)

// allowed OS/Arch combinations
//
//	must be possible value for $GOOS and $GOARCH (https://go.dev/doc/install/source#environment)
//	there must be an OpenTofu release for the respective OS/ARCH (see brokerpakurl.go)
var (
	darwinAmd64  = Platform{Os: "darwin", Arch: "amd64"}
	darwinArm64  = Platform{Os: "darwin", Arch: "arm64"}
	freebsd386   = Platform{Os: "freebsd", Arch: "386"}
	freebsdAmd64 = Platform{Os: "freebsd", Arch: "amd64"}
	freebsdArm   = Platform{Os: "freebsd", Arch: "arm"}
	linux386     = Platform{Os: "linux", Arch: "386"}
	linuxAmd64   = Platform{Os: "linux", Arch: "amd64"}
	linuxArm     = Platform{Os: "linux", Arch: "arm"}
	linuxArm64   = Platform{Os: "linux", Arch: "arm64"}
	openbsd386   = Platform{Os: "openbsd", Arch: "386"}
	openbsdAmd64 = Platform{Os: "openbsd", Arch: "amd64"}
	solarisAmd64 = Platform{Os: "solaris", Arch: "amd64"}
	windows386   = Platform{Os: "windows", Arch: "386"}
	windowsAmd64 = Platform{Os: "windows", Arch: "amd64"}
)

// Validate implements validation.Validatable.
func (p Platform) Validate() (errs *validation.FieldError) {

	if errs := errs.Also(validation.ErrIfBlank(p.Os, "os"), validation.ErrIfBlank(p.Arch, "arch")); errs != nil {
		return errs
	}

	switch p {
	case darwinAmd64, darwinArm64, freebsd386, freebsdAmd64, freebsdArm, linux386, linuxAmd64, linuxArm, linuxArm64, openbsd386, openbsdAmd64, solarisAmd64, windows386, windowsAmd64:
	default:
		return validation.ErrInvalidValue(p, "")
	}

	return nil
}

// String formats the platform as an os/arch pair.
func (p Platform) String() string {
	return fmt.Sprintf("%s/%s", p.Os, p.Arch)
}

// Equals is an equality test between this platform and the other.
func (p Platform) Equals(other Platform) bool {
	return p.String() == other.String()
}

// MatchesCurrent returns true if the platform matches this binary's GOOS/GOARCH combination.
func (p Platform) MatchesCurrent() bool {
	return p.Equals(CurrentPlatform())
}

func (p Platform) Empty() bool {
	return p.Os == "" && p.Arch == ""
}
