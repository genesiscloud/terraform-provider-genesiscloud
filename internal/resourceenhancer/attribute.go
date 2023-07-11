package resourceenhancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Attribute(ctx context.Context, attr schema.Attribute) schema.Attribute {
	switch attr := attr.(type) {
	case schema.BoolAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.Float64Attribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.Int64Attribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.ListAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.MapAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.NumberAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.ObjectAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.SetAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.StringAttribute:
		attr.MarkdownDescription += PlanModifiersMarkdownDescription(ctx, attr.PlanModifiers)
		attr.MarkdownDescription += ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	default:
		return attr
	}
}
