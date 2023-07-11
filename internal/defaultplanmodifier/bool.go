package defaultplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Bool = (*defaultBoolValueAttributePlanModifier)(nil)

// defaultBoolValueAttributePlanModifier specifies a default value (types.Int64) for an attribute.
type defaultBoolValueAttributePlanModifier struct {
	DefaultValue types.Bool
}

// Bool is a helper to instantiate a defaultValueAttributePlanModifier.
func Bool(v bool) planmodifier.Bool {
	return &defaultBoolValueAttributePlanModifier{
		DefaultValue: types.BoolValue(v),
	}
}

func (apm *defaultBoolValueAttributePlanModifier) Description(ctx context.Context) string {
	return apm.MarkdownDescription(ctx)
}

func (apm *defaultBoolValueAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value %q if the attribute is not set.", apm.DefaultValue)
}

func (apm *defaultBoolValueAttributePlanModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, res *planmodifier.BoolResponse) {
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
