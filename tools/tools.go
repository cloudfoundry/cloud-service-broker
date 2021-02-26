// +build tools

package tools

import (
	_ "golang.org/x/tools/cmd/goimports"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
