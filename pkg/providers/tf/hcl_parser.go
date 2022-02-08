package tf

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ReplaceVariable struct {
	Name     string
	Property string
	//Type string
}

func GetSubsumedParameters(tfHCL string, replaceVars []ReplaceVariable) (map[string]interface{}, error) {
	// Need to remove the outputs block from show result
	splitHcl := strings.Split(tfHCL, "Outputs")
	subsumedParameters := make(map[string]interface{})

	// Get the previous values and assign to a map
	parsedConfig, diags := hclwrite.ParseConfig([]byte(splitHcl[0]), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%v", diags.Error())
	}

	for _, block := range parsedConfig.Body().Blocks() {
		if block.Type() == "resource" {
			for _, replaceVar := range replaceVars {
				if stringInSlice(replaceVar.Name, block.Labels()) {
					subsumedParameters[replaceVar.Property] = strings.Trim(strings.TrimSpace(string(block.Body().GetAttribute(replaceVar.Property).Expr().BuildTokens(nil).Bytes())), "\"")
				}
			}
		}

	}
	return subsumedParameters, nil
}

func stringInSlice(subject string, list []string) bool {
	for _, item := range list {
		if item == subject {
			return true
		}
	}
	return false
}
