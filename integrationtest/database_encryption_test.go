package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pivotal-cf/brokerapi/v8/domain"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

var _ = Describe("Database Encryption", func() {
	const (
		provisionParams      = `{"foo":"bar"}`
		bindParams           = `{"baz":"quz"}`
		provisionOutput      = `{"provision_output":"provision output value"}`
		provisionOutputValue = `value = \"provision output value\"`
		bindOutput           = `{"bind_output":"provision output value and bind output value"}`
		bindOutputValue      = `value = \"${var.provision_output} and bind output value\"`
		tfStateKey           = `"tfstate":`
		serviceOfferingGUID  = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID      = "8b52a460-b246-11eb-a8f5-d349948e2480"
	)

	var (
		originalDir         string
		fixturesDir         string
		workDir             string
		brokerPort          int
		brokerUsername      string
		brokerPassword      string
		brokerSession       *Session
		brokerClient        *client.Client
		databaseFile        string
		encryptionKey       string
		serviceInstanceGUID string
		serviceBindingGUID  string
	)

	persistedRequestDetails := func() string {
		db, err := gorm.Open("sqlite3", databaseFile)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		record := models.ProvisionRequestDetails{}
		err = db.Where("service_instance_id = ?", serviceInstanceGUID).First(&record).Error
		Expect(err).NotTo(HaveOccurred())
		return record.RequestDetails
	}

	persistedServiceInstanceDetails := func() string {
		db, err := gorm.Open("sqlite3", databaseFile)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		record := models.ServiceInstanceDetails{}
		err = db.Where("id = ?", serviceInstanceGUID).First(&record).Error
		Expect(err).NotTo(HaveOccurred())
		return record.OtherDetails
	}

	persistedServiceInstanceTerraformWorkspace := func() string {
		db, err := gorm.Open("sqlite3", databaseFile)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		record := models.TerraformDeployment{}
		err = db.Where("id = ?", fmt.Sprintf("tf:%s:", serviceInstanceGUID)).First(&record).Error
		Expect(err).NotTo(HaveOccurred())
		return record.Workspace
	}

	persistedServiceBindingDetails := func() string {
		db, err := gorm.Open("sqlite3", databaseFile)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		record := models.ServiceBindingCredentials{}
		err = db.Where("service_instance_id = ?", serviceInstanceGUID).First(&record).Error
		Expect(err).NotTo(HaveOccurred())
		return record.OtherDetails
	}

	persistedServiceBindingTerraformWorkspace := func() string {
		db, err := gorm.Open("sqlite3", databaseFile)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		record := models.TerraformDeployment{}
		err = db.Where("id = ?", fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID)).First(&record).Error
		Expect(err).NotTo(HaveOccurred())
		return record.Workspace
	}

	createBinding := func() {
		serviceBindingGUID = uuid.New()
		bindResponse := brokerClient.Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
	}

	JustBeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures", "brokerpak-with-fake-provider")

		workDir, err = os.MkdirTemp("", "*-csb-test")
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(workDir)
		Expect(err).NotTo(HaveOccurred())

		files, err := os.ReadDir(fixturesDir)
		Expect(err).ToNot(HaveOccurred())
		for _, file := range files {
			bytes, _ := os.ReadFile(path.Join(fixturesDir, file.Name()))
			os.WriteFile(path.Join(workDir, file.Name()), bytes, 0666)
		}

		buildBrokerpakCommand := exec.Command(csb, "pak", "build")
		session, err := Start(buildBrokerpakCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, time.Minute).Should(Exit(0))

		brokerUsername = uuid.New()
		brokerPassword = uuid.New()
		brokerPort = freePort()
		databaseFile = path.Join(workDir, "databaseFile.dat")
		runBrokerCommand := exec.Command(csb, "serve")
		runBrokerCommand.Env = append(
			os.Environ(),
			"CSB_LISTENER_HOST=localhost",
			"DB_TYPE=sqlite3",
			fmt.Sprintf("EXPERIMENTAL_ENCRYPTION_KEY=%s", encryptionKey),
			fmt.Sprintf("DB_PATH=%s", databaseFile),
			fmt.Sprintf("PORT=%d", brokerPort),
			fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
			fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
		)
		brokerSession, err = Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() bool { return checkAlive(brokerPort) }, 30*time.Second).Should(BeTrue())

		brokerClient, err = client.New(brokerUsername, brokerPassword, "localhost", brokerPort)
		Expect(err).NotTo(HaveOccurred())

		serviceInstanceGUID = uuid.New()
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(provisionParams))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(func() bool {
			lastOperationResponse := brokerClient.LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			var receiver domain.LastOperation
			err = json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("failed"))
			return receiver.State == "succeeded"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
	})

	AfterEach(func() {
		brokerSession.Terminate()

		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("no encryption key is configured", func() {
		BeforeEach(func() {
			encryptionKey = ""
		})

		It("stores sensitive fields in plaintext", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails()).To(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).To(Equal(provisionOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(provisionOutputValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields")
			createBinding()
			Expect(persistedServiceBindingDetails()).To(Equal(bindOutput))
			Expect(persistedServiceBindingTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(bindOutputValue),
				ContainSubstring(tfStateKey),
			))
		})
	})

	When("the encryption key is configured", func() {
		BeforeEach(func() {
			encryptionKey = "one-key-here-with-32-bytes-in-it"
		})

		It("encrypts sensitive fields", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails()).NotTo(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).NotTo(Equal(provisionOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields")
			createBinding()
			Expect(persistedServiceBindingDetails()).NotTo(Equal(bindOutput))
			Expect(persistedServiceBindingTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(bindOutputValue),
				ContainSubstring(tfStateKey),
			))
		})
	})
})

func freePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func checkAlive(port int) bool {
	response, err := http.Head(fmt.Sprintf("http://localhost:%d", port))
	return err == nil && response.StatusCode == http.StatusOK
}

func requestID() string {
	return uuid.New()
}
