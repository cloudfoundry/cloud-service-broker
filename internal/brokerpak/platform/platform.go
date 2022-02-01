package platform

import (
	"fmt"
	"runtime"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

type Platform struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

var _ validation.Validatable = (*Platform)(nil)

func (p Platform) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(p.Os, "os"),
		validation.ErrIfBlank(p.Arch, "arch"),
	)
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

// CurrentPlatform returns the platform defined by GOOS and GOARCH.
func CurrentPlatform() Platform {
	return Platform{Os: runtime.GOOS, Arch: runtime.GOARCH}
}
