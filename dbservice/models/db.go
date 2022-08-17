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

// Package models implements database object models for use with GORM
package models

const (
	// The following operation types are used as part of the OSB process.
	// The types correspond to asynchronous provision/deprovision/update calls
	// and will exist on a ServiceInstanceDetails with an operation ID that can be
	// used to look up the state of an operation.
	ProvisionOperationType   = "provision"
	DeprovisionOperationType = "deprovision"
	UpdateOperationType      = "update"
	UpgradeOperationType     = "upgrade"
	BindOperationType        = "bind"
	UnbindOperationType      = "unbind"
	ClearOperationType       = ""
)

// ServiceBindingCredentials holds credentials returned to the users after
// binding to a service.
type ServiceBindingCredentials ServiceBindingCredentialsV2

// ServiceInstanceDetails holds information about provisioned services.
type ServiceInstanceDetails ServiceInstanceDetailsV3

// ProvisionRequestDetails holds user-defined properties passed to a call
// to provision a service.
type ProvisionRequestDetails ProvisionRequestDetailsV3

// BindRequestDetails holds user-defined properties passed to a call
// to provision a service.
type BindRequestDetails BindRequestDetailsV1

// Migration represents the mgirations table. It holds a monotonically
// increasing number that gets incremented with every database schema revision.
type Migration MigrationV1

// CloudOperation holds information about the status of Google Cloud
// long-running operations.
type CloudOperation CloudOperationV1

// TerraformDeployment holds Terraform state and plan information for resources
// that use that execution system.
type TerraformDeployment TerraformDeploymentV3

// PasswordMetadata contains information about the passwords, but never the
// passwords themselves
type PasswordMetadata PasswordMetadataV1
