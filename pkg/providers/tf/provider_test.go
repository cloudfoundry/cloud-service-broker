package tf_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/spf13/viper"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/varcontext"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type jammingEncryptor struct {
	Error        bool
	EncryptError bool
}

func (d jammingEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if d.EncryptError {
		return nil, fmt.Errorf("error encrypting")
	}
	return []byte("{"), nil
}

func (d jammingEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	var err error
	var result = ciphertext
	if d.Error {
		err = fmt.Errorf("can't decrypt")
	} else {
		result = []byte("{")
	}
	return result, err
}

var _ = FDescribe("WorkspaceUpdator", func() {
	const terraformStateAfterProvision = `
{
  "version": 4,
  "terraform_version": "0.13.7",
  "serial": 17,
  "lineage": "f9da4641-98dd-829a-1406-197a3432356c",
  "outputs": {
    },
  "resources": [
    {
      "mode": "managed",
      "type": "azurerm_sql_database",
      "name": "azure_sql_db",
      "provider": "provider[\"registry.terraform.io/hashicorp/azurerm\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "name": "dbname"
        }
      ]
    }],
`

	var updatedProvisionSettings = tf.TfServiceDefinitionV1Action{
		PlanInputs: []broker.BrokerVariable{
			{
				FieldName: "resourceGroup",
				Type:      broker.JsonTypeString,
				Details:   "The resource group name",
				Required:  true,
			},
		},
		Template: `
				variable resourceGroup {type = string}
	
				resource "random_string" "username" {
				  length = 16
				  special = false
				  number = false
				}

				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = "dbname"
				  resource_group_name = var.resourceGroup
				  administrator_login = random_string.username.result
				}

				output username {value = random_string.username.result}
				`,
		Outputs: []broker.BrokerVariable{
			{
				FieldName: "username",
				Type:      broker.JsonTypeString,
				Details:   "The administrator username",
				Required:  true,
			},
		},
	}
	dummyExecutor := func(ctx context.Context, cmd *exec.Cmd) (wrapper.ExecutionOutput, error) {
		d1 := []byte(terraformStateAfterProvision)
		os.WriteFile(path.Join(cmd.Dir, "terraform.tfstate"), d1, 0644)

		return wrapper.ExecutionOutput{StdOut: "", StdErr: ""}, nil
	}
	workspaceUpdator := tf.WorkspaceUpdator{}

	BeforeEach(func() {
		var err error
		var provider broker.ServiceProvider

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		db_service.DbConnection = db // awful
		defer os.Remove("test.sqlite3")
		models.SetEncryptor(noopencryptor.NoopEncryptor{})
		Expect(err).NotTo(HaveOccurred())
		db.Migrator().CreateTable(models.ServiceInstanceDetails{})
		db.Migrator().CreateTable(models.ServiceBindingCredentials{})
		db.Migrator().CreateTable(models.ProvisionRequestDetails{})
		db.Migrator().CreateTable(models.TerraformDeployment{})

		testLogger := utils.NewLogger("test")
		provisionContext, _ := varcontext.Builder().
			MergeMap(map[string]interface{}{
				"tf_id": "an-instance-id",
			}).
			Build()

		provisionSettings := tf.TfServiceDefinitionV1Action{
			PlanInputs: []broker.BrokerVariable{
				{
					FieldName: "resourceGroup",
					Type:      broker.JsonTypeString,
					Details:   "The resource group name",
					Required:  true,
				},
			},
			UserInputs: []broker.BrokerVariable{
				{
					FieldName: "username",
					Type:      broker.JsonTypeString,
					Details:   "The username to create",
					Required:  true,
				},
			},
			Template: `
				variable resourceGroup {type = string}
				variable username {type = string}
	
				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = "dbname"
				  resource_group_name = var.resourceGroup
				  administrator_login = var.username
				}
				`,
		}
		serviceDefinition := tf.TfServiceDefinitionV1{
			ProvisionSettings: provisionSettings,
		}
		jobRunner := tf.NewTfJobRunnerForProject(map[string]string{})
		jobRunner.Executor = dummyExecutor
		provider = tf.NewTerraformProvider(jobRunner, testLogger, serviceDefinition)
		instanceDetails, _ := provider.Provision(context.TODO(), provisionContext)

		terraformDeploymentFunc := func() string {
			deployment, _ := db_service.GetTerraformDeploymentById(context.TODO(), instanceDetails.OperationId)
			return deployment.LastOperationState
		}
		Eventually(terraformDeploymentFunc).Should(Equal("succeeded"))
		deployment, _ := db_service.GetTerraformDeploymentById(context.TODO(), instanceDetails.OperationId)
		ws, _ := wrapper.DeserializeWorkspace(string(deployment.Workspace))
		Expect(ws.Modules[0].Definition).To(ContainSubstring(`administrator_login = var.username`))
		inputs, _ := ws.Modules[0].Inputs()
		Expect(inputs).To(ConsistOf("resourceGroup", "username"))
		outputs, _ := ws.Modules[0].Outputs()
		Expect(outputs).To(HaveLen(0))
		Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))

	})

	When("dynamic HCL enabled", func() {
		BeforeEach(func() {
			viper.Set("brokerpak.updates.enabled", true)
		})
		When("there's a record in the database", func() {
			It("updates the module definition and variables", func() {
				vc, _ := varcontext.Builder().
					MergeMap(map[string]interface{}{
						"domain": "the domain value",
					}).
					Build()

				workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
				terraformDeployment, _ := db_service.GetTerraformDeploymentById(context.TODO(), "an-instance-id")
				ws, _ := wrapper.DeserializeWorkspace(string(terraformDeployment.Workspace))
				Expect(ws.Modules[0].Definition).To(ContainSubstring(`administrator_login = random_string.username.result`))
				inputs, _ := ws.Modules[0].Inputs()
				Expect(inputs).To(ConsistOf("resourceGroup"))
				outputs, _ := ws.Modules[0].Outputs()
				Expect(outputs).To(ConsistOf("username"))
				Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))

			})
		})
		Context("error scenarios", func() {
			When("there is no record in the database", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id-with-no-workspace")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("record not found"))
				})
			})
			When("cannot get the existing workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()

					models.SetEncryptor(jammingEncryptor{Error: true})
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("can't decrypt"))
				})
			})
			When("cannot deserialize the workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					models.SetEncryptor(jammingEncryptor{Error: false})
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("unexpected end of JSON input"))
				})
			})
			When("cannot create a workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					jammedOperationSettings := tf.TfServiceDefinitionV1Action{
						Template: `
				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = 
				}
				`,
					}
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, jammedOperationSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Invalid expression"))
				})
			})

			When("cannot serialize a workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					jammedOperationSettings := tf.TfServiceDefinitionV1Action{
						Template: `
				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = 
				}
				`,
					}
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, jammedOperationSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Invalid expression"))
				})
			})

			When("cannot set a workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					models.SetEncryptor(jammingEncryptor{EncryptError: true})
					err := workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("unexpected end of JSON input"))
				})
			})

			When("cannot save the workspace", func() {
				It("returns the error", func() {
					vc, _ := varcontext.Builder().Build()
					db, mock, err := sqlmock.New() // mock sql.DB
					mock.ExpectQuery("SELECT*").WillReturnRows(sqlmock.NewRows([]string{"id", "workspace"}).
						AddRow("id", "{}"))
					mock.ExpectBegin()
					mock.ExpectQuery("UPDATE*").WillReturnError(fmt.Errorf("error saving to db"))
					Expect(err).ShouldNot(HaveOccurred())

					dialector := sqlite.Dialector{
						Conn: db,
					}
					dab, err := gorm.Open(dialector, &gorm.Config{})
					db_service.DbConnection = dab
					err = workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("error saving to db"))
				})
			})
		})
	})

	When("dynamic HCL is not enabled", func() {
		BeforeEach(func() {
			viper.Set("brokerpak.updates.enabled", false)
		})
		When("there's a record in the database", func() {
			It("does not update the module definition and variables", func() {
				vc, _ := varcontext.Builder().
					MergeMap(map[string]interface{}{
						"domain": "the domain value",
					}).
					Build()

				workspaceUpdator.UpdateWorkspaceHCL(nil, updatedProvisionSettings, vc, "an-instance-id")
				terraformDeployment, _ := db_service.GetTerraformDeploymentById(context.TODO(), "an-instance-id")
				ws, _ := wrapper.DeserializeWorkspace(string(terraformDeployment.Workspace))
				Expect(ws.Modules[0].Definition).To(ContainSubstring(`administrator_login = var.username`))
				inputs, _ := ws.Modules[0].Inputs()
				Expect(inputs).To(ConsistOf("resourceGroup", "username"))
				outputs, _ := ws.Modules[0].Outputs()
				Expect(outputs).To(HaveLen(0))
				Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))

			})
		})
	})
})
