// +build tools

package tools

import (
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
