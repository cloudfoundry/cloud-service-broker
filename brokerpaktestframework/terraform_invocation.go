package brokerpaktestframework

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

const (
	varsFlatTFDefinition        = "instance.tf.json"
	varsMultiModuleTFDefinition = "terraform.tfvars.json"
)

type TerraformInvocation struct {
	Type string
	dir  string
}

func (i TerraformInvocation) TFVars() (map[string]any, error) {
	filepath, err := i.getTFVarsFilepath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading vars file %w", err)
	}

	if i.isMultiModuleTFDefinition(filepath) {
		output := map[string]any{}
		if err := json.Unmarshal(b, &output); err != nil {
			return nil, fmt.Errorf("error unmarshalling multi-module definition vars file %w", err)
		}
		return output, nil
	}

	type flatModule struct {
		Module struct {
			Instance map[string]any `json:"instance"`
		} `json:"module"`
	}

	var f flatModule
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("error unmarshalling flat definition vars file %w", err)
	}

	return f.Module.Instance, nil
}

func (i TerraformInvocation) getTFVarsFilepath() (string, error) {
	validTFVarsFilesNames := []string{varsMultiModuleTFDefinition, varsFlatTFDefinition}

	for _, filename := range validTFVarsFilesNames {
		p := path.Join(i.dir, filename)
		_, err := os.Stat(p)
		if err != nil && os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("error searching vars file %w", err)
		}

		return p, nil
	}

	return "", fmt.Errorf("vars file not found")
}

func (i TerraformInvocation) isMultiModuleTFDefinition(filename string) bool {
	return strings.Contains(filename, varsMultiModuleTFDefinition)
}
