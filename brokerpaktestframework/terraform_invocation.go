package brokerpaktestframework

import (
	"encoding/json"
	"os"
	"path"
)

type TerraformInvocation struct {
	Type string
	dir  string
}

func (i TerraformInvocation) TFVars() (map[string]interface{}, error) {
	output := map[string]interface{}{}
	file, err := os.Open(path.Join(i.dir, "terraform.tfvars.json"))
	if err != nil {
		return nil, err
	}

	return output, json.NewDecoder(file).Decode(&output)
}
