package tf_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = FDescribe("UpdateWorkspaceHCL", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db_service.RunMigrations(db)).NotTo(HaveOccurred())
	})

	It("works", func() {

		//type TfServiceDefinitionV1 struct {
		//	Version           int                         `yaml:"version"`
		//	Name              string                      `yaml:"name"`
		//	Id                string                      `yaml:"id"`
		//	Description       string                      `yaml:"description"`
		//	DisplayName       string                      `yaml:"display_name"`
		//	ImageUrl          string                      `yaml:"image_url"`
		//	DocumentationUrl  string                      `yaml:"documentation_url"`
		//	SupportUrl        string                      `yaml:"support_url"`
		//	Tags              []string                    `yaml:"tags,flow"`
		//	Plans             []TfServiceDefinitionV1Plan `yaml:"plans"`
		//	ProvisionSettings TfServiceDefinitionV1Action `yaml:"provision"`
		//	BindSettings      TfServiceDefinitionV1Action `yaml:"bind"`
		//	Examples          []broker.ServiceExample     `yaml:"examples"`
		//	PlanUpdateable    bool                        `yaml:"plan_updateable"`
		//
		//	// Internal SHOULD be set to true for Google maintained services.
		//	Internal        bool `yaml:"-"`
		//	RequiredEnvVars []string
		//}
		//serviceDefinition := tf.TfServiceDefinitionV1{
		//	Id:    "some-fake-id",
		//	Name:  "some-fake-definition",
		//	Plans: nil,
		//}
		//provider := tf.NewTerraformProvider(tf.NewTfJobRunnerForProject(map[string]string{}), lager.NewLogger("test"), serviceDefinition)
		//vc := varcontext.VarContext{}
		//tf.WorkspaceUpdator{}.UpdateWorkspaceHCL(nil, serviceDefinition.ProvisionSettings, &vc, "tfId", provider)
	})
})
