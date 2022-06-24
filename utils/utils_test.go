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

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func ExamplePropertyToEnv() {
	env := PropertyToEnv("my.property.key-value")
	fmt.Println(env)

	// Output: GSB_MY_PROPERTY_KEY_VALUE
}

func ExamplePropertyToEnvUnprefixed() {
	env := PropertyToEnvUnprefixed("my.property.key-value")
	fmt.Println(env)

	// Output: MY_PROPERTY_KEY_VALUE
}

func ExampleSetParameter() {
	// Creates an object if none is input
	out, err := SetParameter(nil, "foo", 42)
	fmt.Printf("%s, %v\n", string(out), err)

	// Replaces existing values
	out, err = SetParameter([]byte(`{"replace": "old"}`), "replace", "new")
	fmt.Printf("%s, %v\n", string(out), err)

	// Output: {"foo":42}, <nil>
	// {"replace":"new"}, <nil>
}

func ExampleUnmarshalObjectRemainder() {
	var obj struct {
		A string `json:"a_str"`
		B int
	}

	remainder, err := UnmarshalObjectRemainder([]byte(`{"a_str":"hello", "B": 33, "C": 123}`), &obj)
	fmt.Printf("%s, %v\n", string(remainder), err)

	remainder, err = UnmarshalObjectRemainder([]byte(`{"a_str":"hello", "B": 33}`), &obj)
	fmt.Printf("%s, %v\n", string(remainder), err)

	// Output: {"C":123}, <nil>
	// {}, <nil>
}

func TestSplitNewlineDelimitedList(t *testing.T) {
	cases := map[string]struct {
		Input    string
		Expected []string
	}{
		"none": {
			Input:    ``,
			Expected: nil,
		},
		"single": {
			Input:    `gs://foo/bar`,
			Expected: []string{"gs://foo/bar"},
		},
		"crlf": {
			Input:    "a://foo\r\nb://bar",
			Expected: []string{"a://foo", "b://bar"},
		},
		"trim": {
			Input:    "  a://foo  \n\tb://bar  ",
			Expected: []string{"a://foo", "b://bar"},
		},
		"blank": {
			Input:    "\n\r\r\n\n\t\t\v\n\n\n\n\n \n\n\n \n \n \n\n",
			Expected: nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual := SplitNewlineDelimitedList(tc.Input)

			if !reflect.DeepEqual(tc.Expected, actual) {
				t.Errorf("Expected: %v actual: %v", tc.Expected, actual)
			}
		})
	}
}

func ExampleIndent() {
	weirdText := "First\n\tSecond"
	out := Indent(weirdText, "  ")
	fmt.Println(out == "  First\n  \tSecond")

	// Output: true
}

func ExampleCopyStringMap() {
	m := map[string]string{"a": "one"}
	copy := CopyStringMap(m)
	m["a"] = "two"

	fmt.Println(m["a"])
	fmt.Println(copy["a"])

	// Output: two
	// one
}
