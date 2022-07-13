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

package varcontext

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
)

func TestContextBuilder(t *testing.T) {
	cases := map[string]struct {
		Builder     *ContextBuilder
		Expected    map[string]any
		ErrContains string
	}{
		"an empty context": {
			Builder:     Builder(),
			Expected:    map[string]any{},
			ErrContains: "",
		},

		// MergeMap
		"MergeMap blank okay": {
			Builder:  Builder().MergeMap(map[string]any{}),
			Expected: map[string]any{},
		},
		"MergeMap multi-key": {
			Builder:  Builder().MergeMap(map[string]any{"a": "a", "b": "b"}),
			Expected: map[string]any{"a": "a", "b": "b"},
		},
		"MergeMap overwrite": {
			Builder:  Builder().MergeMap(map[string]any{"a": "a"}).MergeMap(map[string]any{"a": "aaa"}),
			Expected: map[string]any{"a": "aaa"},
		},

		// MergeDefaultWithEval
		"MergeDefaultWithEval no defaults": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "foo"}}),
			Expected: map[string]any{},
		},
		"MergeDefaultWithEval non-string": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "h2g2", Default: 42}}),
			Expected: map[string]any{"h2g2": 42},
		},
		"MergeDefaultWithEval basic-string": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "a", Default: "no-template"}}),
			Expected: map[string]any{"a": "no-template"},
		},
		"MergeDefaultWithEval template string": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "a", Default: "a"}, {Name: "b", Default: "${a}"}}),
			Expected: map[string]any{"a": "a", "b": "a"},
		},
		"MergeDefaultWithEval no-overwrite": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "a", Default: "a"}, {Name: "a", Default: "b", Overwrite: false}}),
			Expected: map[string]any{"a": "a"},
		},
		"MergeDefaultWithEval overwrite": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "a", Default: "a"}, {Name: "a", Default: "b", Overwrite: true}}),
			Expected: map[string]any{"a": "b"},
		},

		"MergeDefaultWithEval object": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "o", Default: `{"foo": "bar"}`, Type: "object"}}),
			Expected: map[string]any{"o": map[string]any{"foo": "bar"}},
		},

		"MergeDefaultWithEval boolean": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "b", Default: `true`, Type: "boolean"}}),
			Expected: map[string]any{"b": true},
		},
		"MergeDefaultWithEval array": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "a", Default: `["a","b","c","d"]`, Type: "array"}}),
			Expected: map[string]any{"a": []any{"a", "b", "c", "d"}},
		},
		"MergeDefaultWithEval number": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "n", Default: `1.234`, Type: "number"}}),
			Expected: map[string]any{"n": 1.234},
		},
		"MergeDefaultWithEval integer": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "i", Default: `1234`, Type: "integer"}}),
			Expected: map[string]any{"i": 1234},
		},
		"MergeDefaultWithEval string": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "s", Default: `1234`, Type: "string"}}),
			Expected: map[string]any{"s": "1234"},
		},
		"MergeDefaultWithEval blank type": {
			Builder:  Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "s", Default: `1234`, Type: ""}}),
			Expected: map[string]any{"s": "1234"},
		},
		"MergeDefaultWithEval bad type": {
			Builder:     Builder().MergeDefaultWithEval([]DefaultVariable{{Name: "s", Default: `1234`, Type: "class"}}),
			ErrContains: "couldn't cast 1234 to class, unknown type",
		},

		// MergeEvalResult
		"MergeEvalResult accumulates context": {
			Builder:  Builder().MergeEvalResult("a", "a", "string").MergeEvalResult("b", "${a}", "string"),
			Expected: map[string]any{"a": "a", "b": "a"},
		},
		"MergeEvalResult errors": {
			Builder:     Builder().MergeEvalResult("a", "${dne}", "string"),
			ErrContains: `couldn't compute the value for "a"`,
		},

		// MergeJSONObject
		"MergeJSONObject blank message": {
			Builder:  Builder().MergeJSONObject(json.RawMessage{}),
			Expected: map[string]any{},
		},
		"MergeJSONObject valid message": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{"a":"a"}`)),
			Expected: map[string]any{"a": "a"},
		},
		"MergeJSONObject invalid message": {
			Builder:     Builder().MergeJSONObject(json.RawMessage(`{{{}}}`)),
			ErrContains: "invalid character '{'",
		},
		"MergeJSONObject merge multiple": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{"foo":"bar"}`)).MergeJSONObject(json.RawMessage(`{"baz":"quz"}`)),
			Expected: map[string]any{"foo": "bar", "baz": "quz"},
		},
		"MergeJSONObject duplicate keys at top level": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{"foo":"bar","baz":"bar"}`)).MergeJSONObject(json.RawMessage(`{"baz":"quz"}`)),
			Expected: map[string]any{"foo": "bar", "baz": "quz"},
		},
		"MergeJSONObject only merges top level key/values": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{"foo":{"bar":"baz","quz":"buz"}}`)).MergeJSONObject(json.RawMessage(`{"foo":{"bar":"quz"}}`)),
			Expected: map[string]any{"foo": map[string]any{"bar": "quz"}},
		},
		"MergeJSONObject merge first empty object": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{}`)).MergeJSONObject(json.RawMessage(`{"baz":"quz"}`)),
			Expected: map[string]any{"baz": "quz"},
		},
		"MergeJSONObject merge second empty object": {
			Builder:  Builder().MergeJSONObject(json.RawMessage(`{"baz":"quz"}`)).MergeJSONObject(json.RawMessage(`{}`)),
			Expected: map[string]any{"baz": "quz"},
		},
		"MergeJSONObject merge JSON non-object": {
			Builder:     Builder().MergeJSONObject(json.RawMessage(`{"baz":"quz"}`)).MergeJSONObject(json.RawMessage(`true`)),
			ErrContains: "json: cannot unmarshal bool into Go value of type map[string]interface {}",
		},

		// MergeStruct
		"MergeStruct without JSON Tags": {
			Builder:  Builder().MergeStruct(struct{ Name string }{Name: "Foo"}),
			Expected: map[string]any{"Name": "Foo"},
		},
		"MergeStruct with JSON Tags": {
			Builder: Builder().MergeStruct(struct {
				Name string `json:"username"`
			}{Name: "Foo"}),
			Expected: map[string]any{"username": "Foo"},
		},

		// constants
		"Basic constants": {
			Builder: Builder().
				SetEvalConstants(map[string]any{"PI": 3.14}).
				MergeEvalResult("out", "${PI}", "string"),
			Expected: map[string]any{"out": "3.14"},
		},
		"User overrides constant": {
			Builder: Builder().
				SetEvalConstants(map[string]any{"PI": 3.14}).
				MergeMap(map[string]any{"PI": 3.2}).      // reassign incorrectly, https://en.wikipedia.org/wiki/Indiana_Pi_Bill
				MergeEvalResult("PI", "${PI}", "string"), // test which PI gets referenced
			Expected: map[string]any{"PI": "3.14"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			vc, err := tc.Builder.Build()

			switch {
			case err == nil && tc.ErrContains == "":
				break
			case err == nil && tc.ErrContains != "":
				t.Fatalf("Got no error when %q was expected", tc.ErrContains)
			case err != nil && tc.ErrContains == "":
				t.Fatalf("Got error %v when none was expected", err)
			case !strings.Contains(err.Error(), tc.ErrContains):
				t.Fatalf("Got error %v, but expected it to contain %q", err, tc.ErrContains)
			}
			if vc == nil && tc.Expected != nil {
				t.Fatalf("Expected: %v, got: %v", tc.Expected, vc)
			}

			if vc != nil && !reflect.DeepEqual(vc.ToMap(), tc.Expected) {
				t.Errorf("Expected: %#v, got: %#v", tc.Expected, vc.ToMap())
			}

		})
	}
}

func ExampleContextBuilder_BuildMap() {
	_, e := Builder().MergeEvalResult("a", "${assert(false, \"failure!\")}", "string").BuildMap()
	fmt.Printf("Error: %v\n", e)

	m, _ := Builder().MergeEvalResult("a", "${1+1}", "string").BuildMap()
	fmt.Printf("Map: %v\n", m)

	//Output: Error: 1 error(s) occurred: couldn't compute the value for "a", template: "${assert(false, \"failure!\")}", assert: assertion failed: failure!
	// Map: map[a:2]
}

func TestDefaultVariable_Validate(t *testing.T) {
	cases := map[string]validation.ValidatableTest{
		"empty": {
			Object: &DefaultVariable{},
			Expect: errors.New("missing field(s): default, name"),
		},
		"bad type": {
			Object: &DefaultVariable{
				Name:    "my-name",
				Default: 123,
				Type:    "stringss",
			},
			Expect: errors.New("field must match '^(|object|boolean|array|number|string|integer)$': type"),
		},
		"good": {
			Object: &DefaultVariable{
				Name:    "my-name",
				Default: 123,
				Type:    "string",
			},
			Expect: nil,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.Assert(t)
		})
	}
}
