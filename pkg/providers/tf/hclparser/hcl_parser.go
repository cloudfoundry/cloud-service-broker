// Package hclparser is used to parse HCL (Hashicorp Configuration Language) that Terraform is written in
package hclparser

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ExtractVariable struct {
	FieldToRead  string
	FieldToWrite string
}

func GetParameters(tfHCL string, parameters []ExtractVariable) (map[string]any, error) {
	splitHcl := strings.Split(tfHCL, "Outputs")
	parsedConfig, diags := hclwrite.ParseConfig([]byte(splitHcl[0]), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error parsing subsumed HCL file: %v", diags.Error())
	}

	subsumedParameters := make(map[string]any)
	for _, block := range parsedConfig.Body().Blocks() {
		if block.Type() == "resource" {
			for _, param := range parameters {
				resourceName, attributeName := splitResource(param.FieldToRead)
				if resourceName == strings.Join(block.Labels(), ".") {
					if val, ok := getAttribute(block, attributeName); ok {
						subsumedParameters[param.FieldToWrite] = val
					}
				}
			}
		}
	}

	var notFoundParameters []string
	if len(subsumedParameters) != len(parameters) {
		for _, rv := range parameters {
			if _, ok := subsumedParameters[rv.FieldToWrite]; !ok {
				notFoundParameters = append(notFoundParameters, rv.FieldToRead)
			}
		}
		return nil, fmt.Errorf("cannot find required subsumed values for fields: %s", strings.Join(notFoundParameters, ", "))
	}

	return subsumedParameters, nil
}

func splitResource(resource string) (string, string) {
	lastInd := strings.LastIndex(resource, ".")
	return resource[:lastInd], resource[lastInd+1:]
}

func getAttribute(block *hclwrite.Block, attribute string) (string, bool) {
	attr := block.Body().GetAttribute(attribute)
	if attr == nil {
		return "", false
	}
	return strings.Trim(strings.TrimSpace(string(attr.Expr().BuildTokens(nil).Bytes())), "\""), true
}
