package wrapper_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Terraform Commands", func() {
	Context("initCommand", func() {
		It("calls init with the plugin directory", func() {
			initCommand := wrapper.NewInitCommand("plugindir")
			Expect(initCommand.Command()).To(Equal([]string{"init", "-plugin-dir=plugindir", "-no-color"}))
		})
	})

	Context("init012Command", func() {
		It("calls init with the plugin directory", func() {
			initCommand := wrapper.NewInit012Command("plugindir")
			Expect(initCommand.Command()).To(Equal([]string{"init", "-plugin-dir=plugindir", "-get-plugins=false", "-no-color"}))
		})
	})
})
