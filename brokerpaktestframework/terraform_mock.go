package brokerpaktestframework

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/onsi/gomega/gexec"
)

func NewTerraformMock() (TerraformMock, error) {
	dir, err := os.MkdirTemp("", "invocation_store")
	if err != nil {
		return TerraformMock{}, err
	}
	build, err := gexec.Build("github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework/mock-binary/terraform", "-ldflags", fmt.Sprintf("-X 'main.InvocationStore=%s'", dir))
	if err != nil {
		return TerraformMock{}, err
	}
	return TerraformMock{Binary: build, invocationStore: dir, Version: "1.1.4"}, nil
}

type TerraformMock struct {
	Binary          string
	invocationStore string
	Version         string
}

func (p TerraformMock) ApplyInvocations() ([]TerraformInvocation, error) {
	invocations, err := p.Invocations()
	if err != nil {
		return nil, err
	}
	filteredInovocations := []TerraformInvocation{}
	for _, invocation := range invocations {
		if invocation.Type == "apply" {
			filteredInovocations = append(filteredInovocations, invocation)
		}
	}
	return filteredInovocations, nil
}

func (p TerraformMock) Invocations() ([]TerraformInvocation, error) {
	fileInfo, err := os.ReadDir(p.invocationStore)
	if err != nil {
		return nil, err
	}
	var invocations []TerraformInvocation

	for _, file := range fileInfo {
		parts := strings.Split(file.Name(), "-")
		invocations = append(invocations, TerraformInvocation{Type: parts[0], dir: path.Join(p.invocationStore, file.Name())})
	}
	return invocations, nil
}

func (p TerraformMock) Reset() error {
	dir, err := os.ReadDir(p.invocationStore)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err := os.RemoveAll(path.Join(p.invocationStore, d.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (p TerraformMock) FirstTerraformInvocationVars() (map[string]interface{}, error) {
	invocations, err := p.ApplyInvocations()
	if err != nil {
		return nil, err
	}
	if len(invocations) != 1 {
		return nil, fmt.Errorf("unexpected number of invocations, acutal: %d expected %d", len(invocations), 1)
	}

	vars, err := invocations[0].TFVars()
	if err != nil {
		return nil, err
	}
	return vars, nil
}

func (p TerraformMock) setTFStateFile(state workspace.Tfstate) error {
	file, err := os.Create(path.Join(p.invocationStore, "mock_tf_state.json"))
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(state)
}

type TFStateValue struct {
	Name  string
	Type  string
	Value interface{}
}

// SetTFState set the Terraform State in a JSON file.
func (p TerraformMock) SetTFState(values []TFStateValue) error {
	var outputs = make(map[string]struct {
		Type  string      `json:"type"`
		Value interface{} `json:"value"`
	})
	for _, value := range values {
		outputs[value.Name] = struct {
			Type  string      `json:"type"`
			Value interface{} `json:"value"`
		}{
			Type:  value.Type,
			Value: value.Value,
		}
	}

	return p.setTFStateFile(workspace.Tfstate{
		Version:          4,
		TerraformVersion: p.Version,
		Outputs:          outputs})
}

// ReturnTFState set the Terraform State in a JSON file.
// Deprecated: due to the introduction of a new name that provides a more accurate meaning.
// We use parallel change to not break backwards compatibility.
// To set the Terraform State use the TerraformMock.SetTFState method.
func (p TerraformMock) ReturnTFState(values []TFStateValue) error {
	return p.SetTFState(values)
}
