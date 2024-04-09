package command_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Terraform Commands", func() {
	Context("initCommand", func() {
		It("calls init with the plugin directory", func() {
			initCommand := command.NewInit("plugindir")
			Expect(initCommand.Command()).To(Equal([]string{"init", "-plugin-dir=plugindir", "-no-color"}))
			Expect(initCommand.Env()).To(BeEmpty())
		})
	})

	Context("apply", func() {
		It("calls apply with the right options", func() {
			apply := command.NewApply()
			Expect(apply.Command()).To(Equal([]string{"apply", "-auto-approve", "-no-color"}))
			Expect(apply.Env()).To(BeEmpty())
		})
	})

	Context("Destroy", func() {
		It("calls destroy with the right options", func() {
			destroy := command.NewDestroy()
			Expect(destroy.Command()).To(Equal([]string{"destroy", "-auto-approve", "-no-color"}))
			Expect(destroy.Env()).To(BeEmpty())
		})
	})

	Context("Show", func() {
		It("calls show with the right env variables", func() {
			show := command.NewShow()
			Expect(show.Command()).To(Equal([]string{"show", "-no-color"}))
			Expect(show.Env()).To(Equal([]string{"OPENTOFU_STATEFILE_PROVIDER_ADDRESS_TRANSLATION=0"}))
		})
	})

	Context("Plan", func() {
		It("calls show with the right env variables", func() {
			plan := command.NewPlan()
			Expect(plan.Command()).To(Equal([]string{"plan", "-no-color"}))
			Expect(plan.Env()).To(BeEmpty())
		})
	})
})
