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

package models

import (
	"encoding/json"
)

const (
	// The following operation types are used as part of the OSB process.
	// The types correspond to asynchronous provision/deprovision/update calls
	// and will exist on a ServiceInstanceDetails with an operation ID that can be
	// used to look up the state of an operation.
	ProvisionOperationType   = "provision"
	DeprovisionOperationType = "deprovision"
	UpdateOperationType      = "update"
	ClearOperationType       = ""
)

var encryptorInstance Encryptor = nil

func SetEncryptor(encryptor Encryptor) {
	encryptorInstance = encryptor
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/fake_encryption.go . Encryptor
type Encryptor interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}

// ServiceBindingCredentials holds credentials returned to the users after
// binding to a service.
type ServiceBindingCredentials ServiceBindingCredentialsV1

// SetOtherDetails marshals the value passed in into a JSON string and sets
// OtherDetails to it if marshalling was successful.
func (sbc *ServiceBindingCredentials) SetOtherDetails(toSet interface{}) error {
	out, err := json.Marshal(toSet)
	if err != nil {
		return err
	}

	encryptedDetails, err := encryptorInstance.Encrypt(out)
	if err != nil {
		return err
	}

	sbc.OtherDetails = string(encryptedDetails)
	return nil
}

// GetOtherDetails returns and unmarshalls the OtherDetails field into the given
// struct. An empty OtherDetails field does not get unmarshalled and does not error.
func (sbc ServiceBindingCredentials) GetOtherDetails(v interface{}) error {
	if sbc.OtherDetails == "" {
		return nil
	}

	decryptedDetails, err := encryptorInstance.Decrypt(sbc.OtherDetails)
	if err != nil {
		return err
	}

	return json.Unmarshal(decryptedDetails, v)
}

// ServiceInstanceDetails holds information about provisioned services.
type ServiceInstanceDetails ServiceInstanceDetailsV2

// SetOtherDetails marshals the value passed in into a JSON string and sets
// OtherDetails to it if marshalling was successful.
func (si *ServiceInstanceDetails) SetOtherDetails(toSet interface{}) error {
	out, err := json.Marshal(toSet)
	if err != nil {
		return err
	}

	encryptedDetails, err := encryptorInstance.Encrypt(out)
	if err != nil {
		return err
	}

	si.OtherDetails = string(encryptedDetails)
	return nil
}

// GetOtherDetails returns and unmarshalls the OtherDetails field into the given
// struct. An empty OtherDetails field does not get unmarshalled and does not error.
func (si ServiceInstanceDetails) GetOtherDetails(v interface{}) error {
	if si.OtherDetails == "" {
		return nil
	}

	decryptedDetails, err := encryptorInstance.Decrypt(si.OtherDetails)
	if err != nil {
		return err
	}

	return json.Unmarshal(decryptedDetails, v)
}

// ProvisionRequestDetails holds user-defined properties passed to a call
// to provision a service.
type ProvisionRequestDetails ProvisionRequestDetailsV1

func (pr *ProvisionRequestDetails) SetRequestDetails(rawMessage json.RawMessage) error {
	encryptedDetails, err := encryptorInstance.Encrypt(rawMessage)
	if err != nil {
		return err
	}

	pr.RequestDetails = string(encryptedDetails)
	return nil
}

func (pr ProvisionRequestDetails) GetRequestDetails() (json.RawMessage, error) {
	decryptedDetails, err := encryptorInstance.Decrypt(pr.RequestDetails)
	if err != nil {
		return nil, err
	}

	return decryptedDetails, nil
}

// Migration represents the mgirations table. It holds a monotonically
// increasing number that gets incremented with every database schema revision.
type Migration MigrationV1

// CloudOperation holds information about the status of Google Cloud
// long-running operations.
type CloudOperation CloudOperationV1

// TerraformDeployment holds Terraform state and plan information for resources
// that use that execution system.
type TerraformDeployment TerraformDeploymentV2

func (t *TerraformDeployment) SetWorkspace(value string) error {
	encrypted, err := encryptorInstance.Encrypt([]byte(value))
	if err != nil {
		return err
	}

	t.Workspace = encrypted
	return nil
}

func (t *TerraformDeployment) GetWorkspace() (string, error) {
	decrypted, err := encryptorInstance.Decrypt(t.Workspace)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// PasswordMetadata contains information about the passwords, but never the
// passwords themselves
type PasswordMetadata PasswordMetadataV1
