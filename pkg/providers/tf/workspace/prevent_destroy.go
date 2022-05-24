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
	for i, m := range workspace.Modules {
		var err error
		workspace.Modules[i].Definition, err = removePreventDestroy(m.Definition)
		if err != nil {
			return fmt.Errorf("HCL parse error for module %d definition: %w", i, err)
		}
		for k, v := range workspace.Modules[i].Definitions {
			workspace.Modules[i].Definitions[k], err = removePreventDestroy(v)
			if err != nil {
				return fmt.Errorf("HCL parse error for module %d definition %q: %w", i, k, err)
			}
		}
	}
	return nil
}

func removePreventDestroy(input string) (string, error) {
	f, diags := hclwrite.ParseConfig([]byte(input), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return "", diags
	}

	for _, resource := range f.Body().Blocks() {
		if resource.Type() == resourceIdentifier {
			if lifecycle := resource.Body().FirstMatchingBlock(lifecycleIdentifier, nil); lifecycle != nil {
				if pd := lifecycle.Body().GetAttribute(preventDestroyIdentifier); pd != nil {
					lifecycle.Body().SetAttributeValue(preventDestroyIdentifier, cty.False)
				}
			}
		}
	}

	return string(f.Bytes()), nil
}
