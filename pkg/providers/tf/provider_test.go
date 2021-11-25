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
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models/fakes"

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

var _ = Describe("WorkspaceUpdater", func() {
	var (
		workspaceUpdator tf.WorkspaceUpdater
		vc               *varcontext.VarContext
	)

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

	updatedProvisionSettings := tf.TfServiceDefinitionV1Action{
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

	setUpProvider := func(serviceDefinition tf.TfServiceDefinitionV1) broker.ServiceProvider {
		testLogger := utils.NewLogger("test")
		jobRunner := tf.NewTfJobRunnerForProject(map[string]string{})
		jobRunner.Executor = dummyExecutor
		return tf.NewTerraformProvider(jobRunner, testLogger, serviceDefinition)
	}

	pollOperationSucceeded := func(operationId string) func() string {
		return func() string {
			deployment, err := db_service.GetTerraformDeploymentById(context.TODO(), operationId)
			Expect(err).NotTo(HaveOccurred())
			return deployment.LastOperationState
		}
	}

	expectModuleToBeInitialHCL := func(ws *wrapper.TerraformWorkspace) {
		Expect(ws.Modules[0].Definition).To(ContainSubstring(`administrator_login = var.username`))
		inputs, err := ws.Modules[0].Inputs()
		Expect(err).NotTo(HaveOccurred())
		Expect(inputs).To(ConsistOf("resourceGroup", "username"))
		outputs, err := ws.Modules[0].Outputs()
		Expect(err).NotTo(HaveOccurred())
		Expect(outputs).To(HaveLen(0))
	}

	getTerraformWorkspace := func(operationId string) *wrapper.TerraformWorkspace {
		deployment, err := db_service.GetTerraformDeploymentById(context.TODO(), operationId)
		Expect(err).NotTo(HaveOccurred())
		ws, err := wrapper.DeserializeWorkspace(string(deployment.Workspace))
		Expect(err).ToNot(HaveOccurred())
		return ws
	}

	BeforeEach(func() {
		var err error
		var provider broker.ServiceProvider

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		db_service.DbConnection = db // awful
		models.SetEncryptor(noopencryptor.NoopEncryptor{})
		Expect(err).NotTo(HaveOccurred())
		db.Migrator().CreateTable(models.ServiceInstanceDetails{})
		db.Migrator().CreateTable(models.ServiceBindingCredentials{})
		db.Migrator().CreateTable(models.ProvisionRequestDetails{})
		db.Migrator().CreateTable(models.TerraformDeployment{})

		vc, err = varcontext.Builder().Build()
		Expect(err).NotTo(HaveOccurred())

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

		provider = setUpProvider(tf.TfServiceDefinitionV1{
			ProvisionSettings: provisionSettings,
		})

		provisionContext, err := varcontext.Builder().
			MergeMap(map[string]interface{}{
				"tf_id": "an-instance-id",
			}).
			Build()
		Expect(err).NotTo(HaveOccurred())
		instanceDetails, err := provider.Provision(context.TODO(), provisionContext)
		Expect(err).NotTo(HaveOccurred())

		Eventually(pollOperationSucceeded(instanceDetails.OperationId)).Should(Equal("succeeded"))
		ws := getTerraformWorkspace(instanceDetails.OperationId)
		expectModuleToBeInitialHCL(ws)
		Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))

	})

	When("brokerpak updates enabled", func() {
		BeforeEach(func() {
			viper.Set("brokerpak.updates.enabled", true)

		})

		When("there's a record in the database", func() {
			It("updates the module definition and variables", func() {
				workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")

				ws := getTerraformWorkspace("an-instance-id")
				Expect(ws.Modules[0].Definition).To(ContainSubstring(`administrator_login = random_string.username.result`))
				inputs, err := ws.Modules[0].Inputs()
				Expect(err).NotTo(HaveOccurred())
				Expect(inputs).To(ConsistOf("resourceGroup"))
				outputs, err := ws.Modules[0].Outputs()
				Expect(err).NotTo(HaveOccurred())
				Expect(outputs).To(ConsistOf("username"))
				Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))
			})
		})

		Context("error scenarios", func() {
			When("there is no record in the database", func() {
				It("returns the error", func() {
					err := workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id-with-no-workspace")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("record not found"))
				})
			})

			When("cannot get the existing workspace", func() {
				It("returns the error", func() {
					encryptor := fakes.FakeEncryptor{}
					encryptor.DecryptReturns(nil, fmt.Errorf("can't decrypt"))
					models.SetEncryptor(&encryptor)

					err := workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("can't decrypt"))
				})
			})

			When("cannot deserialize the workspace", func() {
				It("returns the error", func() {
					encryptor := fakes.FakeEncryptor{}
					encryptor.DecryptReturns([]byte("{"), nil)
					models.SetEncryptor(&encryptor)

					err := workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("unexpected end of JSON input"))
				})
			})

			When("cannot create a workspace", func() {
				It("returns the error", func() {
					jammedOperationSettings := tf.TfServiceDefinitionV1Action{
						Template: `
				resource "azurerm_mssql_database" "azure_sql_db" {
				  name                = 
				}
				`,
					}
					err := workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), jammedOperationSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Invalid expression"))
				})
			})

			When("cannot set a workspace", func() {
				It("returns the error", func() {
					encryptor := fakes.FakeEncryptor{}
					encryptor.DecryptReturns([]byte("{}"), nil)
					encryptor.EncryptReturns(nil, fmt.Errorf("error encrypting"))
					models.SetEncryptor(&encryptor)
					err := workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("error encrypting"))
				})
			})

			When("cannot save the workspace", func() {
				It("returns the error", func() {
					db, mock, err := sqlmock.New()
					Expect(err).ShouldNot(HaveOccurred())
					mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("4.5.6"))
					mock.ExpectQuery("SELECT*").WillReturnRows(sqlmock.NewRows([]string{"id", "workspace"}).
						AddRow("id", "{}"))
					mock.ExpectBegin()
					mock.ExpectQuery("UPDATE*").WillReturnError(fmt.Errorf("error saving to db"))
					dialector := sqlite.Dialector{
						Conn: db,
					}
					mockDbConnection, err := gorm.Open(dialector, &gorm.Config{})
					Expect(err).ShouldNot(HaveOccurred())
					db_service.DbConnection = mockDbConnection

					err = workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("error saving to db"))
				})
			})
		})
	})

	When("brokerpak updates are not enabled", func() {
		BeforeEach(func() {
			viper.Set("brokerpak.updates.enabled", false)
		})

		When("there's a record in the database", func() {
			It("does not update the module definition and variables", func() {
				workspaceUpdator.UpdateWorkspaceHCL(context.TODO(), updatedProvisionSettings, vc, "an-instance-id")

				ws := getTerraformWorkspace("an-instance-id")
				expectModuleToBeInitialHCL(ws)
				Expect(string(ws.State)).To(Equal(terraformStateAfterProvision))
			})
		})
	})
})
