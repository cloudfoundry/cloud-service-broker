package workspace

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	resourceIdentifier       = "resource"
	lifecycleIdentifier      = "lifecycle"
	preventDestroyIdentifier = "prevent_destroy"
)

func (workspace *TerraformWorkspace) RemovePreventDestroy() error {
	for index, module := range workspace.Modules {
		var err error
		workspace.Modules[index].Definition, err = removePreventDestroy(module.Definition)
		if err != nil {
			return fmt.Errorf("HCL parse error for module %d definition: %w", index, err)
		}
		for definitionName, definitionContents := range workspace.Modules[index].Definitions {
			workspace.Modules[index].Definitions[definitionName], err = removePreventDestroy(definitionContents)
			if err != nil {
				return fmt.Errorf("HCL parse error for module %d definition %q: %w", index, definitionName, err)
			}
		}
	}
	return nil
}

func removePreventDestroy(input string) (string, error) {
	file, diags := hclwrite.ParseConfig([]byte(input), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", diags
	}

	for _, resource := range file.Body().Blocks() {
		if resource.Type() == resourceIdentifier {
			if lifecycle := resource.Body().FirstMatchingBlock(lifecycleIdentifier, nil); lifecycle != nil {
				if pd := lifecycle.Body().GetAttribute(preventDestroyIdentifier); pd != nil {
					lifecycle.Body().SetAttributeValue(preventDestroyIdentifier, cty.False)
				}
			}
		}
	}

	return string(file.Bytes()), nil
}
