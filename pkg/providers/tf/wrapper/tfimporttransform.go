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
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// ParameterMapping mapping for tf variable to service parameter
type ParameterMapping struct {
	TfVariable string `yaml:"tf_variable"`
	ParameterName string `yaml:"parameter_name"`
}

// TfTransformer terraform transformation 
type TfTransformer struct {
	ParameterMappings []ParameterMapping `json:"parameter_mappings"`
	ParametersToRemove []string			 `json:"parameters_to_remove"`
}

func braceCount(str string, count int) int {
	return count + strings.Count(str, "{") - strings.Count(str, "}")
}

// CleanTf removes ttf.ParametersToRemove from tf string
func (ttf *TfTransformer) CleanTf(tf string) string {
	resource := regexp.MustCompile(`resource "(.*)" "(.*)"` )
	value := regexp.MustCompile(`^[\s]*([^\s]*)[\s]*=[\s]*(.*)[\s]*$`)
	block := regexp.MustCompile(`^[\s]*([^\s]*)[\s]*{[\s]*$`)		
	depth := 0
	blockStack := make([]string, 64)
	scanner := bufio.NewScanner(strings.NewReader(tf))
	buffer := bytes.Buffer{}

	for scanner.Scan() {
		skipLine := false
		line := scanner.Text()
		depth = braceCount(line, depth)
		
		if res := resource.FindStringSubmatch(line); res != nil {
			blockStack[depth] = fmt.Sprintf("%s.%s", res[1], res[2])
		} else if res = value.FindStringSubmatch(line); res != nil {
			for _, removal := range ttf.ParametersToRemove {
				if fmt.Sprintf("%s.%s", blockStack[depth], res[1]) == removal {
					skipLine = true
					break
				}
			}
		} else if res := block.FindStringSubmatch(line); res != nil {
			blockStack[depth] = fmt.Sprintf("%s.%s", blockStack[depth-1], res[1])
		}
		if !skipLine {
			buffer.WriteString(fmt.Sprintf("%s\n", line))
		}		
	}
	return buffer.String()
}

func (ttf *TfTransformer) captureParameterValues(tf string) (map[string]string, error) {
	parameterValues := make(map[string]string)

	for _, mapping := range ttf.ParameterMappings {
		reBlock := regexp.MustCompile(fmt.Sprintf(`(?m)%s[\s]*=[\s]+({[\s\S.]*?})`, mapping.TfVariable))
		reSimple := regexp.MustCompile(fmt.Sprintf(`(?m)%s[\s]*=[\s"]*(.*[^"\s])`, mapping.TfVariable))

		if res := reBlock.FindAllStringSubmatch(tf, -1); len(res) > 0 {
			//parameterValues[mapping.ParameterName] = res[0][1]
		} else if res := reSimple.FindAllStringSubmatch(tf, -1); len(res) > 0 {
			if strings.HasPrefix(mapping.ParameterName, "var.") {
				parameterValues[strings.TrimPrefix(mapping.ParameterName, "var.")] = res[0][1]
			} else if strings.HasPrefix(mapping.ParameterName, "local.") {
				parameterValues[strings.TrimPrefix(mapping.ParameterName, "local.")] = res[0][1]
			}
		}
	}
	
	return parameterValues, nil
}

func (ttf *TfTransformer) replaceParameters(tf string) string {	
	for _, mapping := range ttf.ParameterMappings {
		reBlock := regexp.MustCompile(fmt.Sprintf(`(?m)%s[\s]*=[\s]+{[\s\S.]*?}`, mapping.TfVariable))
		tf = reBlock.ReplaceAllString(tf, fmt.Sprintf("%s = %s", mapping.TfVariable, mapping.ParameterName))
		reSimple := regexp.MustCompile(fmt.Sprintf(`(?m)%s[\s]*=.*$`, mapping.TfVariable))
		tf = reSimple.ReplaceAllString(tf, fmt.Sprintf("%s = %s", mapping.TfVariable, mapping.ParameterName))
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