// Package packer knows how to create a brokerpak given a manifest,
// a source directory and a destination.
package packer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/brokerpakurl"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/fetcher"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/stream"
	"github.com/hashicorp/go-getter"
)

const manifestName = "manifest.yml"

func Pack(m *manifest.Manifest, base, dest string) error {
	// NOTE: we use "log" rather than Lager because this is used by the CLI and
	// needs to be human-readable rather than JSON.
	switch base {
	case "":
		log.Printf("Packing brokerpak version %q with CSB version %q...\n", m.Version, utils.Version)
	default:
		log.Printf("Packing %q version %q with CSB version %q...\n", base, m.Version, utils.Version)
	}

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
	packSource := func(source, name string) error {
		if source == "" {
			return nil
		}
		destination := filepath.Join(tmp, "src", name+".zip")

		log.Println("\t", source, "->", destination)
		return fetcher.FetchArchive(source, destination)
	}

	for _, resource := range m.TerraformVersions {
		if err := packSource(resource.Source, "terraform"); err != nil {
			return err
		}
	}
	for _, resource := range m.TerraformProviders {
		if err := packSource(resource.Source, resource.Name); err != nil {
			return err
		}
	}
	for _, resource := range m.Binaries {
		if err := packSource(resource.Source, resource.Name); err != nil {
			return err
		}
	}

	return nil
}

func packBinaries(m *manifest.Manifest, tmp string) error {
	for _, platform := range m.Platforms {
		p := filepath.Join(tmp, "bin", platform.Os, platform.Arch)

		for _, resource := range m.TerraformVersions {
			log.Println("\t", brokerpakurl.URL("terraform", resource.Version.String(), resource.URLTemplate, platform), "->", filepath.Join(p, resource.Version.String()))
			if err := getter.GetAny(filepath.Join(p, resource.Version.String()), brokerpakurl.URL("terraform", resource.Version.String(), resource.URLTemplate, platform)); err != nil {
				return err
			}
		}
		for _, resource := range m.TerraformProviders {
			log.Println("\t", brokerpakurl.URL(resource.Name, resource.Version.String(), resource.URLTemplate, platform), "->", p)
			if err := getter.GetAny(p, brokerpakurl.URL(resource.Name, resource.Version.String(), resource.URLTemplate, platform)); err != nil {
				return err
			}
		}
		for _, resource := range m.Binaries {
			log.Println("\t", brokerpakurl.URL(resource.Name, resource.Version, resource.URLTemplate, platform), "->", p)
			if err := getter.GetAny(p, brokerpakurl.URL(resource.Name, resource.Version, resource.URLTemplate, platform)); err != nil {
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
	data, err := manifestCopy.Serialize()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, manifestName), data, 0600); err != nil {
		return err
	}

	return nil
}

func clearRefs(sd *tf.TfServiceDefinitionV1Action) {
	sd.TemplateRef = ""
	sd.TemplateRefs = make(map[string]string)
}
