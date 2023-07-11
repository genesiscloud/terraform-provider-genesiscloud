package providerenhancer

import (
	"context"

	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func Attribute(ctx context.Context, attr schema.Attribute) schema.Attribute {
	switch attr := attr.(type) {
	case schema.BoolAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.Float64Attribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.Int64Attribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.ListAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.MapAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.NumberAttribute:

		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.ObjectAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.SetAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	case schema.StringAttribute:
		attr.MarkdownDescription += resourceenhancer.ValidatorsMarkdownDescription(ctx, attr.Validators)
		return attr

	default:
		return attr
	}
}
