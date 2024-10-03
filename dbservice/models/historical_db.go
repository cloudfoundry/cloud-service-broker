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
	"time"

	"gorm.io/gorm"
)

// This file contains versioned models for the database, so we
// can do proper tracking through gorm.
//
// If you need to change a model you MUST make a copy here and update the
// reference to the new model in db.go and add a migration path in the
// dbservice package.

// ServiceBindingCredentialsV1 holds credentials returned to the users after
// binding to a service.
type ServiceBindingCredentialsV1 struct {
	gorm.Model

	OtherDetails string `gorm:"type:text"`

	ServiceID         string
	ServiceInstanceID string
	BindingID         string
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceBindingCredentialsV1) TableName() string {
	return "service_binding_credentials"
}

// ServiceBindingCredentialsV2 holds credentials returned to the users after
// binding to a service.
type ServiceBindingCredentialsV2 struct {
	gorm.Model

	OtherDetails []byte `gorm:"type:blob"`

	ServiceID         string
	ServiceInstanceID string
	BindingID         string
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceBindingCredentialsV2) TableName() string {
	return "service_binding_credentials"
}

// ServiceInstanceDetailsV1 holds information about provisioned services.
type ServiceInstanceDetailsV1 struct {
	ID        string `gorm:"primary_key;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Name         string
	Location     string
	URL          string
	OtherDetails string `gorm:"type:text"`

	ServiceID        string
	PlanID           string
	SpaceGUID        string
	OrganizationGUID string
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceInstanceDetailsV1) TableName() string {
	return "service_instance_details"
}

// ServiceInstanceDetailsV2 holds information about provisioned services.
type ServiceInstanceDetailsV2 struct {
	ID        string `gorm:"primary_key;type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Name         string
	Location     string
	URL          string
	OtherDetails string `gorm:"type:text"`

	ServiceID        string
	PlanID           string
	SpaceGUID        string
	OrganizationGUID string

	// OperationType holds a string corresponding to what kind of operation
	// OperationID is referencing. The object is "locked" for editing if
	// an operation is pending.
	OperationType string

	// OperationID holds a string referencing an operation specific to a broker.
	// Operations in GCP all have a unique ID.
	// The OperationID will be cleared after a successful operation.
	// This string MAY be sent to users and MUST NOT leak confidential information.
	OperationID string `gorm:"type:varchar(1024)"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceInstanceDetailsV2) TableName() string {
	return "service_instance_details"
}

// ServiceInstanceDetailsV3 holds information about provisioned services.
type ServiceInstanceDetailsV3 struct {
	ID        string `gorm:"primary_key;type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Name         string
	Location     string
	URL          string
	OtherDetails []byte `gorm:"type:blob"`

	ServiceID        string
	PlanID           string
	SpaceGUID        string
	OrganizationGUID string

	// OperationType holds a string corresponding to what kind of operation
	// OperationID is referencing. The object is "locked" for editing if
	// an operation is pending.
	OperationType string

	// OperationID holds a string referencing an operation specific to a broker.
	// Operations in GCP all have a unique ID.
	// The OperationID will be cleared after a successful operation.
	// This string MAY be sent to users and MUST NOT leak confidential information.
	OperationID string `gorm:"type:varchar(1024)"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceInstanceDetailsV3) TableName() string {
	return "service_instance_details"
}

// ServiceInstanceDetailsV3 holds information about provisioned services.
type ServiceInstanceDetailsV4 struct {
	ID        string `gorm:"primary_key;type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Name         string
	Location     string
	URL          string
	OtherDetails []byte `gorm:"type:blob"`

	ServiceID        string
	PlanID           string
	SpaceGUID        string
	OrganizationGUID string
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ServiceInstanceDetailsV4) TableName() string {
	return "service_instance_details"
}

// ProvisionRequestDetailsV1 holds user-defined properties passed to a call
// to provision a service.
type ProvisionRequestDetailsV1 struct {
	gorm.Model

	ServiceInstanceID string `gorm:"uniqueIndex"`
	// is a json.Marshal of models.ProvisionDetails
	RequestDetails string
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ProvisionRequestDetailsV1) TableName() string {
	return "provision_request_details"
}

// ProvisionRequestDetailsV2 holds user-defined properties passed to a call
// to provision a service.
type ProvisionRequestDetailsV2 struct {
	gorm.Model

	ServiceInstanceID string

	// is a json.Marshal of models.ProvisionDetails
	RequestDetails string `gorm:"type:text"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ProvisionRequestDetailsV2) TableName() string {
	return "provision_request_details"
}

// ProvisionRequestDetailsV3 holds user-defined properties passed to a call
// to provision a service.
type ProvisionRequestDetailsV3 struct {
	gorm.Model

	ServiceInstanceID string

	// is a json.Marshal of models.ProvisionDetails
	RequestDetails []byte `gorm:"type:blob"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (ProvisionRequestDetailsV3) TableName() string {
	return "provision_request_details"
}

// BindRequestDetailsV1 holds user-defined properties passed to a call
// to provision a service.
type BindRequestDetailsV1 struct {
	gorm.Model

	ServiceBindingID  string `gorm:"unique"`
	ServiceInstanceID string

	// is a json.Marshal of models.BindDetails
	RequestDetails []byte `gorm:"type:blob"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (BindRequestDetailsV1) TableName() string {
	return "bind_request_details"
}

// MigrationV1 represents the migrations table. It holds a monotonically
// increasing number that gets incremented with every database schema revision.
type MigrationV1 struct {
	gorm.Model

	MigrationID int `gorm:"type:int(10)"`
}

// TableName returns a consistent table name for gorm so
// multiple structs from different versions of the database all operate on the
// same table.
func (MigrationV1) TableName() string {
	return "migrations"
}

// CloudOperationV1 holds information about the status of Google Cloud
// long-running operations.
// As-of version 4.1.0, this table is no longer necessary.
type CloudOperationV1 struct {
	gorm.Model

	Name          string
	Status        string
	OperationType string
	ErrorMessage  string `gorm:"type:text"`
	InsertTime    string
	StartTime     string
	TargetID      string
	TargetLink    string

	ServiceID         string
	ServiceInstanceID string
}

// TableName returns a consistent table name for gorm so
// multiple structs from different versions of the database all operate on the
// same table.
func (CloudOperationV1) TableName() string {
	return "cloud_operations"
}

// PlanDetailsV1 is a table that was deprecated in favor of using Environment
// variables. It only remains for ORM migrations and the ability for existing
// users to export their plans.
type PlanDetailsV1 struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	ServiceID string
	Name      string
	Features  string `gorm:"type:text"`
}

// TableName returns a consistent table name for gorm so
// multiple structs from different versions of the database all operate on the
// same table.
func (PlanDetailsV1) TableName() string {
	return "plan_details"
}

// TerraformDeploymentV1 describes the state of a Terraform resource deployment.
type TerraformDeploymentV1 struct {
	ID        string `gorm:"primary_key;type:varchar(1024)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Workspace contains a JSON serialized version of the Terraform workspace.
	Workspace string `gorm:"type:mediumtext"`

	// LastOperationType describes the last operation being performed on the resource.
	LastOperationType string

	// LastOperationState holds one of the following strings "in progress", "succeeded", "failed".
	// These mirror the OSB API.
	LastOperationState string

	// LastOperationMessage is a description that can be passed back to the user.
	LastOperationMessage string `gorm:"type:text"`
}

// TableName returns a consistent table name for gorm so
// multiple structs from different versions of the database all operate on the
// same table.
func (TerraformDeploymentV1) TableName() string {
	return "terraform_deployments"
}

// TerraformDeploymentV2 expands the size of the Workspace column to handle deployments where the
// Terraform workspace is greater than 64K. (mediumtext allows for workspaces up
// to 16384K.)
type TerraformDeploymentV2 struct {
	ID        string `gorm:"primary_key;type:varchar(1024)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Workspace contains a JSON serialized version of the Terraform workspace.
	Workspace string `gorm:"type:mediumtext"`

	// LastOperationType describes the last operation being performed on the resource.
	LastOperationType string

	// LastOperationState holds one of the following strings "in progress", "succeeded", "failed".
	// These mirror the OSB API.
	LastOperationState string

	// LastOperationMessage is a description that can be passed back to the user.
	LastOperationMessage string `gorm:"type:text"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (TerraformDeploymentV2) TableName() string {
	return "terraform_deployments"
}

// TerraformDeploymentV3 converts workspace type from mediumtext to mediumblob
type TerraformDeploymentV3 struct {
	ID        string `gorm:"primary_key;type:varchar(1024)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Workspace contains a JSON serialized version of the Terraform workspace.
	Workspace []byte `gorm:"type:mediumblob"`

	// LastOperationType describes the last operation being performed on the resource.
	LastOperationType string

	// LastOperationState holds one of the following strings "in progress", "succeeded", "failed".
	// These mirror the OSB API.
	LastOperationState string

	// LastOperationMessage is a description that can be passed back to the user.
	LastOperationMessage string `gorm:"type:text"`
}

// TableName returns a consistent table name for
// gorm so multiple structs from different versions of the database all operate
// on the same table.
func (TerraformDeploymentV3) TableName() string {
	return "terraform_deployments"
}

// PasswordMetadataV1 contains information about the passwords, but never the
// passwords themselves
type PasswordMetadataV1 struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Label   string `gorm:"index;unique;not null"`
	Salt    []byte `gorm:"type:blob;not null"`
	Canary  []byte `gorm:"type:blob;not null"`
	Primary bool
}

func (PasswordMetadataV1) TableName() string {
	return "password_metadata"
}
