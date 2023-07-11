package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/defaultplanmodifier"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                     = &SecurityGroupResource{}
	_ resource.ResourceWithConfigure        = &SecurityGroupResource{}
	_ resource.ResourceWithImportState      = &SecurityGroupResource{}
	_ resource.ResourceWithConfigValidators = &SecurityGroupResource{}
)

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

// SecurityGroupResource defines the resource implementation.
type SecurityGroupResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

type SecurityGroupResourceModel struct {
	SecurityGroupModel

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Security group resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this security group was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description for the security group.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the security group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the security group.",
				Required:            true,
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The region identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllComputeV1Regions)...),
				},
			}),
			"rules": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"direction": resourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "The direction of the rule.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(sliceStringify(genesiscloud.AllComputeV1SecurityGroupRuleDirections)...),
							},
						}),
						"port_range_max": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
							MarkdownDescription: "The maximum port number of the rule.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						}),
						"port_range_min": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
							MarkdownDescription: "The minimum port number of the rule.",
							Optional:            true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						}),
						"protocol": resourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "The protocol of the rule.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(sliceStringify(genesiscloud.AllComputeV1SecurityGroupRuleProtocols)...),
							},
						}),
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The security group status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *SecurityGroupResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// Improvement: Add additional rule validation
	}
}

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Create)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := genesiscloud.CreateSecurityGroupJSONRequestBody{}

	body.Description = pointer(data.Description.ValueString())
	body.Name = data.Name.ValueString()
	body.Region = pointer(genesiscloud.ComputeV1Region(data.Region.ValueString()))

	for _, rule := range data.Rules {
		var portRangeMax, portRangeMin *int

		if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
			portRangeMax = pointer(int(rule.PortRangeMax.ValueInt64()))
		}

		if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
			portRangeMin = pointer(int(rule.PortRangeMin.ValueInt64()))
		}

		body.Rules = append(body.Rules, genesiscloud.ComputeV1SecurityGroupRule{
			Direction:    genesiscloud.ComputeV1SecurityGroupRuleDirection(rule.Direction.ValueString()),
			PortRangeMax: portRangeMax,
			PortRangeMin: portRangeMin,
			Protocol:     genesiscloud.ComputeV1SecurityGroupRuleProtocol(rule.Protocol.ValueString()),
		})
	}

	response, err := r.client.CreateSecurityGroupWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create security_group", err))
		return
	}

	securityGroupResponse := response.JSON201
	if securityGroupResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create security_group", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &securityGroupResponse.SecurityGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a security group resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	securityGroupId := securityGroupResponse.SecurityGroup.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling security_group", err))
			return
		}

		tflog.Trace(ctx, "polling a security group resource")

		response, err := r.client.GetSecurityGroupWithResponse(ctx, securityGroupId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling security_group", err))
			return
		}

		securityGroupResponse := response.JSON200
		if securityGroupResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling security_group", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := securityGroupResponse.SecurityGroup.Status
		if status == "created" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &securityGroupResponse.SecurityGroup)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling security_group", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	securityGroupId := data.Id.ValueString()

	response, err := r.client.GetSecurityGroupWithResponse(ctx, securityGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read security_group", err))
		return
	}

	securityGroupResponse := response.JSON200
	if securityGroupResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read security_group", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &securityGroupResponse.SecurityGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a security group resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Update)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := genesiscloud.UpdateSecurityGroupJSONRequestBody{}

	body.Description = pointer(data.Description.ValueString())
	body.Name = data.Name.ValueString()
	for _, rule := range data.Rules {
		var portRangeMax, portRangeMin *int

		if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
			portRangeMax = pointer(int(rule.PortRangeMax.ValueInt64()))
		}

		if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
			portRangeMin = pointer(int(rule.PortRangeMin.ValueInt64()))
		}

		body.Rules = append(body.Rules, genesiscloud.ComputeV1SecurityGroupRule{
			Direction:    genesiscloud.ComputeV1SecurityGroupRuleDirection(rule.Direction.ValueString()),
			PortRangeMax: portRangeMax,
			PortRangeMin: portRangeMin,
			Protocol:     genesiscloud.ComputeV1SecurityGroupRuleProtocol(rule.Protocol.ValueString()),
		})
	}

	securityGroupId := data.Id.ValueString()

	response, err := r.client.UpdateSecurityGroupWithResponse(ctx, securityGroupId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update security_group", err))
		return
	}

	securityGroupResponse := response.JSON200
	if securityGroupResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update security_group", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &securityGroupResponse.SecurityGroup)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a security group resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling security_group", err))
			return
		}

		tflog.Trace(ctx, "polling a security group resource")

		response, err := r.client.GetSecurityGroupWithResponse(ctx, securityGroupId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling security_group", err))
			return
		}

		securityGroupResponse := response.JSON200
		if securityGroupResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling security_group", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := securityGroupResponse.SecurityGroup.Status
		if status == "created" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &securityGroupResponse.SecurityGroup)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling security_group", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Delete)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	securityGroupId := data.Id.ValueString()

	response, err := r.client.DeleteSecurityGroupWithResponse(ctx, securityGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete security_group", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete security_group", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling security_group", err))
			return
		}

		tflog.Trace(ctx, "polling a security group resource")

		response, err := r.client.GetSecurityGroupWithResponse(ctx, securityGroupId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling security_group", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
