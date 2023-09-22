package resourceenhancer

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// TODO: Upstream better quotes and remove this
func patchMarkdown(s string) string {
	s = strings.ReplaceAll(s, `"\"`, `"`)
	s = strings.ReplaceAll(s, `\""`, `"`)
	return s
}

func PlanModifiersMarkdownDescription[T planmodifier.Describer](ctx context.Context, mods []T) string {
	response := ""

	for _, mod := range mods {
		desc := mod.MarkdownDescription(ctx)
		if desc == "Once set, the value of this attribute in state will not change." {
			continue
		}

		response += "\n  - " + patchMarkdown(desc)
	}

	return response
}

func ValidatorsMarkdownDescription[T validator.Describer](ctx context.Context, validators []T) string {
	response := ""

	for _, validator := range validators {
		// TODO: Check existing dot and "The" vs "It" etc.
		response += "\n  - The " + patchMarkdown(validator.MarkdownDescription(ctx)) + "."
	}

	return response
}
