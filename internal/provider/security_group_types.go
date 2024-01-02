package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityGroupRuleModel struct {
	// Direction The direction of the rule.
	Direction types.String `tfsdk:"direction"`

	// PortRangeMax The maximum port number of the rule.
	PortRangeMax types.Int64 `tfsdk:"port_range_max"`

	// PortRangeMin The minimum port number of the rule.
	PortRangeMin types.Int64 `tfsdk:"port_range_min"`

	// Protocol The protocol of the rule.
	Protocol types.String `tfsdk:"protocol"`
}

type SecurityGroupResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Description The human-readable description for the security group.
	Description types.String `tfsdk:"description"`

	// Id The unique ID of the security group.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the security group.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// Rules The security group rules.
	Rules []SecurityGroupRuleModel `tfsdk:"rules"`

	// Status The security group status.
	Status types.String `tfsdk:"status"`

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *SecurityGroupResourceModel) PopulateFromClientResponse(ctx context.Context, securityGroup *genesiscloud.SecurityGroup) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(securityGroup.CreatedAt.Format(time.RFC3339))
	data.Description = types.StringValue(securityGroup.Description)
	data.Id = types.StringValue(securityGroup.Id)
	data.Name = types.StringValue(securityGroup.Name)
	data.Region = types.StringValue(string(securityGroup.Region))

	data.Rules = nil
	for _, rule := range securityGroup.Rules {
		portRangeMax := types.Int64Null()
		portRangeMin := types.Int64Null()

		if rule.PortRangeMax != nil {
			portRangeMax = types.Int64Value(int64(*rule.PortRangeMax))
		}

		if rule.PortRangeMin != nil {
			portRangeMin = types.Int64Value(int64(*rule.PortRangeMin))
		}

		data.Rules = append(data.Rules, SecurityGroupRuleModel{
			Direction:    types.StringValue(string(rule.Direction)),
			PortRangeMax: portRangeMax,
			PortRangeMin: portRangeMin,
			Protocol:     types.StringValue(string(rule.Protocol)),
		})
	}

	data.Status = types.StringValue(string(securityGroup.Status))

	return
}
