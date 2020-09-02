// Copyright 2020 the VMware, Inc.
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

package wrapper

import (
	"fmt"
	"regexp"
)


// ParameterMapping mapping for tf variable to service parameter
type ParameterMapping struct {
	TfVariable string `yaml:"tf_variable"`
	ParameterName string `yaml:"parameter_name"`
}

// TfTransformer terraform transformation 
type TfTransformer struct {
	ParameterMappings []ParameterMapping
	ParametersToRemove []string
}

// CleanTf removes ttf.ParametersToRemove from tf string
func (ttf *TfTransformer) CleanTf(tf string) string {
	removals := ""

	for n, removal := range ttf.ParametersToRemove {
		if n == 0 {
			removals = removal
		} else {
			removals += fmt.Sprintf("|%s", removal)
		}
	}
	re := regexp.MustCompile(fmt.Sprintf(`(?m)^[\s]+(%s)[\s]*=.*$`, removals))
	return re.ReplaceAllString(tf, "")	
}

func (ttf *TfTransformer) captureParameterValues(tf string) (map[string]string, error) {
	parameterValues := make(map[string]string)

	for _, mapping := range ttf.ParameterMappings {
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^[\s]+%s[\s]*=[\s"]*(.*[^"\s])`, mapping.TfVariable))
		res := re.FindAllStringSubmatch(tf, -1)
		if len(res) > 1 {
			return parameterValues, fmt.Errorf("Found more than one tf parameter %s in %s", mapping.TfVariable, tf )
		} else if len(res) > 0 {
			parameterValues[mapping.ParameterName] = res[0][1]
		}
	}
	
	return parameterValues, nil
}

func (ttf *TfTransformer) replaceParameters(tf string) string {	
	for _, mapping := range ttf.ParameterMappings {
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^[\s]+(%s)[\s]*=.*$`, mapping.TfVariable))
		tf = re.ReplaceAllString(tf, fmt.Sprintf("%s = var.%s", mapping.TfVariable, mapping.ParameterName))
	}
	
	return tf
}

// ReplaceParametersInTf replaces ttf.ParameterMappings in tf
func (ttf *TfTransformer) ReplaceParametersInTf(tf string) (string, map[string]string, error) {
	parameterValues, err := ttf.captureParameterValues(tf)

	if err == nil {
		tf = ttf.replaceParameters(tf) 
	}
	return tf, parameterValues, err
}