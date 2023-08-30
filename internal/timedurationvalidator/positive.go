package timedurationvalidator

// Adapted from https://github.com/hashicorp/terraform-plugin-framework-timeouts/blob/3b8726d5a4e0c7204ce6b0529f2b37db9f8551c4/internal/validators/timeduration.go

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = timeDurationPositiveValidator{}

// timeDurationPositiveValidator validates that a string Attribute's value is parseable as time.Duration.
type timeDurationPositiveValidator struct {
}

// Description describes the validation in plain text formatting.
func (validator timeDurationPositiveValidator) Description(_ context.Context) string {
	return `string must be a positive [time duration](https://pkg.go.dev/time#ParseDuration), for example "10s"`
}

// MarkdownDescription describes the validation in Markdown formatting.
func (validator timeDurationPositiveValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// ValidateString performs the validation.
func (validator timeDurationPositiveValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	s := req.ConfigValue

	if s.IsUnknown() || s.IsNull() {
		return
	}

	duration, err := time.ParseDuration(s.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid Attribute Value Time Duration",
			fmt.Sprintf("%q %s", s.ValueString(), validator.Description(ctx))),
		)
		return
	}

	if duration < 0 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Time Duration Must Be Positive",
			fmt.Sprintf("%q %s", s.ValueString(), validator.Description(ctx))),
		)
		return
	}
}

// Positive returns an AttributeValidator which ensures that any configured
// attribute value:
//
//   - Is parseable as time duration.
//   - Is positive.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func Positive() validator.String {
	return timeDurationPositiveValidator{}
}
