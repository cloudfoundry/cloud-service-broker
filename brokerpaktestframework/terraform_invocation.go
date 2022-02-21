package brokerpaktestframework

import (
	"encoding/json"
	"io"
	"os"
	"path"
)

type TerraformInvocation struct {
	Type string
	dir  string
}

func (i TerraformInvocation) TFVars() (map[string]interface{}, error) {
	tfVarsContents, err := i.TFVarsContents()
	if err != nil {
		return nil, err
	}
	output := map[string]interface{}{}

	return output, json.Unmarshal([]byte(tfVarsContents), &output)
}

func (i TerraformInvocation) TFVarsContents() (string, error) {
	file, err := os.Open(path.Join(i.dir, "terraform.tfvars.json"))
	if err != nil {
		return "", err
	}
	all, err := io.ReadAll(file)
	return string(all), err

}
