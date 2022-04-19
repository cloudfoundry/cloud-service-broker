package workspace

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdditionalStateFiles", func() {
	Context("saving", func() {
		It("saves *.pem files", func() {
			workdir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(workdir, "terraform.tfstate"), []byte("fake-terraform-state"), 0755)).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(workdir, "fake.pem"), []byte("fake-certificate"), 0600)).NotTo(HaveOccurred())
			w := &TerraformWorkspace{dir: workdir}

			Expect(w.saveState()).NotTo(HaveOccurred())

			Expect(w.State).To(Equal([]byte("fake-terraform-state")))
			Expect(w.AdditionalState).To(Equal(map[string][]byte{
				"fake.pem": []byte("fake-certificate"),
			}))
		})
	})

	Context("restoring", func() {
		It("restores files", func() {
			workdir := GinkgoT().TempDir()
			w := &TerraformWorkspace{
				State: []byte("fake-terraform-state"),
				AdditionalState: map[string][]byte{
					"fake.pem": []byte("fake-certificate"),
				},
				dir: workdir,
			}

			Expect(w.restoreState()).NotTo(HaveOccurred())

			Expect(filepath.Join(workdir, "terraform.tfstate")).To(BeAnExistingFile())
			Expect(filepath.Join(workdir, "fake.pem")).To(BeAnExistingFile())
		})
	})
})
