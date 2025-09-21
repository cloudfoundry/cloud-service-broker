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

// Package utils contains various utils that do various things.
// It is deprecated and any future utils should be added as an appropriately named sub-package.
// Over time this package should be emptied.
package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"regexp"
	"strings"

	"code.cloudfoundry.org/lager/v3"
)

const EnvironmentVarPrefix = "gsb"

var (
	PropertyToEnvReplacer = strings.NewReplacer(".", "_", "-", "_")

	// InvalidLabelChars encodes that GCP labels only support alphanumeric,
	// dash and underscore characters in keys and values.
	InvalidLabelChars = regexp.MustCompile("[^a-zA-Z0-9_-]+")
)

// PrettyPrintOrExit writes a JSON serialized version of the content to stdout.
// If a failure occurs during marshaling, the error is logged along with a
// formatted version of the object and the program exits with a failure status.
func PrettyPrintOrExit(content any) {
	err := prettyPrint(content)

	if err != nil {
		log.Fatalf("Could not format results: %s, results were: %+v", err, content)
	}
}

func prettyPrint(content any) error {
	prettyResults, err := json.MarshalIndent(content, "", "    ")
	if err == nil {
		fmt.Println(string(prettyResults))
	}

	return err
}

// PropertyToEnv converts a Viper configuration property name into an
// environment variable prefixed with EnvironmentVarPrefix
func PropertyToEnv(propertyName string) string {
	return PropertyToEnvUnprefixed(EnvironmentVarPrefix + "." + propertyName)
}

// PropertyToEnvUnprefixed converts a Viper configuration property name into an
// environment variable using PropertyToEnvReplacer
func PropertyToEnvUnprefixed(propertyName string) string {
	return PropertyToEnvReplacer.Replace(strings.ToUpper(propertyName))
}

// SetParameter sets a value on a JSON raw message and returns a modified
// version with the value set
func SetParameter(input json.RawMessage, key string, value any) (json.RawMessage, error) {
	params := make(map[string]any)

	if len(input) != 0 {
		err := json.Unmarshal(input, &params)
		if err != nil {
			return nil, err
		}
	}

	params[key] = value

	return json.Marshal(params)
}

// UnmarshalObjectRemainder unmarshals an object into v and returns the
// remaining key/value pairs as a JSON string by doing a set difference.
func UnmarshalObjectRemainder(data []byte, v any) ([]byte, error) {
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return jsonDiff(data, encoded)
}

func jsonDiff(superset, subset json.RawMessage) ([]byte, error) {
	usedKeys := make(map[string]json.RawMessage)
	if err := json.Unmarshal(subset, &usedKeys); err != nil {
		return nil, err
	}

	allKeys := make(map[string]json.RawMessage)
	if err := json.Unmarshal(superset, &allKeys); err != nil {
		return nil, err
	}

	remainder := make(map[string]json.RawMessage)
	for key, value := range allKeys {
		if _, ok := usedKeys[key]; !ok {
			remainder[key] = value
		}
	}

	return json.Marshal(remainder)
}

// SingleLineErrorFormatter creates a single line error string from an array of errors.
func SingleLineErrorFormatter(es []error) string {
	points := make([]string, len(es))
	for i, err := range es {
		points[i] = err.Error()
	}

	return fmt.Sprintf("%d error(s) occurred: %s", len(es), strings.Join(points, "; "))
}

// NewLogger creates a new lager.Logger with the given name that has correct
// writing settings.
func NewLogger(name string) lager.Logger {
	logger := lager.NewLogger(name)
	logLevel := lager.INFO // default

	// Can use environment variable CSB_LOG_LEVEL to set the level.
	// If the value is invalid, we ignore it.
	if level, ok := os.LookupEnv("CSB_LOG_LEVEL"); ok {
		parsedLevel, err := lager.LogLevelFromString(level)
		if err == nil {
			logLevel = parsedLevel
		}
	}

	// The GSB_DEBUG environment variable is the long-standing way
	// to enable debug logging, and it overrides CSB_LOG_LEVEL
	if _, debug := os.LookupEnv("GSB_DEBUG"); debug {
		logLevel = lager.DEBUG
	}

	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, logLevel))

	return logger
}

// SplitNewlineDelimitedList splits a list of newline delimited items and trims
// any leading or trailing whitespace from them.
func SplitNewlineDelimitedList(paksText string) []string {
	var out []string
	for pak := range strings.SplitSeq(paksText, "\n") {
		pakURL := strings.TrimSpace(pak)
		if pakURL != "" {
			out = append(out, pakURL)
		}
	}

	return out
}

// Indent indents every line of the given text with the given string.
func Indent(text, by string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		lines[i] = by + line
	}

	return strings.Join(lines, "\n")
}

// CopyStringMap makes a copy of the given map.
func CopyStringMap(m map[string]string) map[string]string {
	out := make(map[string]string)

	maps.Copy(out, m)

	return out
}
