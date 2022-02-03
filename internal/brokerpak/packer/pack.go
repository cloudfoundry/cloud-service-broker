// Package packer knows how to create a brokerpak given a manifest,
// a source directory and a destination.
package packer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/brokerpakurl"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/fetcher"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
	"github.com/hashicorp/go-getter"
)

const manifestName = "manifest.yml"

func Pack(m *manifest.Manifest, base, dest string) error {
	// NOTE: we use "log" rather than Lager because this is used by the CLI and
	// needs to be human readable rather than JSON.
	log.Println("Packing...")

	dir, err := os.MkdirTemp("", "brokerpak")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up
	log.Println("Using temp directory:", dir)

	log.Println("Packing sources...")
	if err := packSources(m, dir); err != nil {
		return err
	}

	log.Println("Packing binaries...")
	if err := packBinaries(m, dir); err != nil {
		return err
	}

	log.Println("Packing definitions...")
	if err := packDefinitions(m, dir, base); err != nil {
		return err
	}

	log.Println("Creating archive:", dest)
	return zippy.Archive(dir, dest)
}

func packSources(m *manifest.Manifest, tmp string) error {
	for _, resource := range m.TerraformResources {
		if resource.Source == "" {
			continue
		}

		destination := filepath.Join(tmp, "src", resource.Name+".zip")

		log.Println("\t", resource.Source, "->", destination)
		if err := fetcher.FetchArchive(resource.Source, destination); err != nil {
			return err
		}
	}

	return nil
}

func packBinaries(m *manifest.Manifest, tmp string) error {
	for _, platform := range m.Platforms {
		for _, resource := range m.TerraformResources {
			p := filepath.Join(tmp, "bin", platform.Os, platform.Arch)

			if resource.Name == "terraform" {
				p = filepath.Join(p, resource.Version)
			}

			log.Println("\t", brokerpakurl.URL(resource, platform), "->", p)
			if err := getter.GetAny(p, brokerpakurl.URL(resource, platform)); err != nil {
				return err
			}
		}
	}

	return nil
}

func packDefinitions(m *manifest.Manifest, tmp, base string) error {
	// users can place definitions in any directory structure they like, even
	// above the current directory so we standardize their location and names
	// for the zip to avoid collisions
	//
	// provision and bind templates are loaded from any template ref and packed inline
	manifestCopy := *m

	var (
		servicePaths []string
		catalog      tf.TfCatalogDefinitionV1
	)

	for i, sd := range m.ServiceDefinitions {

		defn := &tf.TfServiceDefinitionV1{}
		if err := stream.Copy(stream.FromFile(base, sd), stream.ToYaml(defn)); err != nil {
			return fmt.Errorf("couldn't parse %s: %v", sd, err)
		}

		if err := defn.ProvisionSettings.LoadTemplate(base); err != nil {
			return fmt.Errorf("couldn't load provision template %s: %v", defn.ProvisionSettings.TemplateRef, err)
		}

		if err := defn.BindSettings.LoadTemplate(base); err != nil {
			return fmt.Errorf("couldn't load bind template %s: %v", defn.BindSettings.TemplateRef, err)
		}

		clearRefs(&defn.ProvisionSettings)
		clearRefs(&defn.BindSettings)

		packedName := fmt.Sprintf("service%d-%s.yml", i, defn.Name)
		log.Printf("\t%s/%s -> %s/definitions/%s\n", base, sd, tmp, packedName)
		if err := stream.Copy(stream.FromYaml(defn), stream.ToFile(tmp, "definitions", packedName)); err != nil {
			return err
		}

		servicePaths = append(servicePaths, "definitions/"+packedName)
		catalog = append(catalog, defn)
	}

	if err := catalog.Validate(); err != nil {
		return err
	}

	manifestCopy.ServiceDefinitions = servicePaths

	return stream.Copy(stream.FromYaml(manifestCopy), stream.ToFile(tmp, manifestName))
}

func clearRefs(sd *tf.TfServiceDefinitionV1Action) {
	sd.TemplateRef = ""
	sd.TemplateRefs = make(map[string]string)
}
