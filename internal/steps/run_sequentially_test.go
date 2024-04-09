package steps_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/steps"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunSequentially", func() {
	It("does the things in sequence", func() {
		step := 0

		err := steps.RunSequentially(
			func() error {
				if step != 0 {
					Fail("not at step 0")
				}
				step++
				return nil
			},
			func() error {
				if step != 1 {
					Fail("not at step 1")
				}
				step++
				return nil
			},
			func() error {
				if step != 2 {
					Fail("not at step 2")
				}
				step++
				return nil
			},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(step).To(Equal(3))
	})

	It("returns errors", func() {
		step := 0

		err := steps.RunSequentially(
			func() error {
				if step != 0 {
					Fail("not at step 0")
				}
				step++
				return nil
			},
			func() error {
				if step != 1 {
					Fail("not at step 1")
				}
				step++
				return fmt.Errorf("boom")
			},
			func() error {
				if step != 2 {
					Fail("not at step 2")
				}
				step++
				return nil
			},
		)

		Expect(err).To(MatchError("boom"))
		Expect(step).To(Equal(2))
	})
})
