// Package testmatchers implements custom test matchers
package testmatchers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/onsi/gomega/types"
)

func MatchDirectoryContents(path string) types.GomegaMatcher {
	return &directoryContentsMatcher{expected: path}
}

type directoryContentsMatcher struct {
	expected string
}

func (m *directoryContentsMatcher) Match(actual any) (bool, error) {
	a, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("actual must be a string")
	}

	err := recursiveEquals(m.expected, a)
	switch err.(type) {
	case nil:
		return true, nil
	default:
		return false, nil
	}
}

func (m *directoryContentsMatcher) FailureMessage(actual any) string {
	a, ok := actual.(string)
	if !ok {
		return "actual must be a string"
	}

	err := recursiveEquals(m.expected, a)
	return fmt.Sprintf("Expected %s to match directory %s, but: %s", m.expected, a, err.Error())
}

func (m *directoryContentsMatcher) NegatedFailureMessage(actual any) string {
	a, ok := actual.(string)
	if !ok {
		return "actual must be a string"
	}

	return fmt.Sprintf("Expected %s to not to match directory %s, but it did", m.expected, a)
}

func recursiveEquals(expected, actual string) error {
	ehandle, err := os.Open(expected)
	if err != nil {
		return err
	}
	ahandle, err := os.Open(actual)
	if err != nil {
		return err
	}

	einfo, err := ehandle.Stat()
	if err != nil {
		return err
	}
	ainfo, err := ahandle.Stat()
	if err != nil {
		return err
	}

	if einfo.IsDir() {
		if !ainfo.IsDir() {
			return fmt.Errorf("expected %s is dir but actual %s is not", expected, actual)
		}

		const all = 0
		acontents, err := ahandle.ReadDir(all)
		if err != nil {
			return err
		}
		econtents, err := ehandle.ReadDir(all)
		if err != nil {
			return err
		}

		actualNames := make(map[string]struct{})
		for _, dir := range acontents {
			actualNames[dir.Name()] = struct{}{}
		}

		for _, dir := range econtents {
			if _, ok := actualNames[dir.Name()]; !ok {
				return fmt.Errorf("expected file does not exist: %s", path.Join(expected, dir.Name()))
			}
			err := recursiveEquals(path.Join(expected, dir.Name()), path.Join(actual, dir.Name()))
			if err != nil {
				return err
			}
			delete(actualNames, dir.Name())
		}

		if len(actualNames) != 0 {
			var extraneous []string
			for name := range actualNames {
				extraneous = append(extraneous, name)
			}
			return fmt.Errorf("unxpected extra files %q in %s", strings.Join(extraneous, ", "), actual)
		}
	} else {
		if ainfo.IsDir() {
			return fmt.Errorf("expected %s is file but actual %s is not", expected, actual)
		}

		econtents, err := io.ReadAll(ehandle)
		if err != nil {
			return err
		}

		acontents, err := io.ReadAll(ahandle)
		if err != nil {
			return err
		}

		if !bytes.Equal(econtents, acontents) {
			return fmt.Errorf("expected file %s to have the same contents as %s and it does not", actual, expected)
		}
	}

	return nil
}
