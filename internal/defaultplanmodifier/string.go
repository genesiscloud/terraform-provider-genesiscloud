package defaultplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = (*defaultStringValueAttributePlanModifier)(nil)

// defaultStringValueAttributePlanModifier specifies a default value (types.Int64) for an attribute.
type defaultStringValueAttributePlanModifier struct {
	DefaultValue types.String
}

// String is a helper to instantiate a defaultValueAttributePlanModifier.
func String(v string) planmodifier.String {
	return &defaultStringValueAttributePlanModifier{
		DefaultValue: types.StringValue(v),
	}
}

func (apm *defaultStringValueAttributePlanModifier) Description(ctx context.Context) string {
	return apm.MarkdownDescription(ctx)
}

func (apm *defaultStringValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value %q if the attribute is not set.", apm.DefaultValue)
}

func (apm *defaultStringValueAttributePlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, res *planmodifier.StringResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	res.PlanValue = apm.DefaultValue
}
