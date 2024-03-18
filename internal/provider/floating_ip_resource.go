package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/defaultplanmodifier"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	_ resource.Resource                = &FloatingIPResource{}
	_ resource.ResourceWithConfigure   = &FloatingIPResource{}
	_ resource.ResourceWithImportState = &FloatingIPResource{}
)

func NewFloatingIPResource() resource.Resource {
	return &FloatingIPResource{}
}

// FloatingIPResource defines the resource implementation.
type FloatingIPResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *FloatingIPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (r *FloatingIPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "floating IP resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this floating IP was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the floating IP.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the floating IP.",
				Required:            true,
			}),
			"updated_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this image was last updated in RFC 3339.",
				Computed:            true,
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The region identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllRegions)...),
				},
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description set for the floating IP.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The floating IP status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"version": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The version of the floating IP.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(string(genesiscloud.CreateFloatingIPJSONBodyVersionIpv4)),
					// stringvalidator.OneOf(sliceStringify(genesiscloud.AllFloatingIPVersions)...),
				},
			}),
			"ip_address": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The IP address of the floating IP.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"is_public": resourceenhancer.Attribute(ctx, schema.BoolAttribute{
				MarkdownDescription: "Whether the floating IP is public or private.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					defaultplanmodifier.Bool(true),
				},
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *FloatingIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FloatingIPResourceModel

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

	body := genesiscloud.CreateFloatingIPJSONRequestBody{}

	body.Name = data.Name.ValueString()
	body.Region = genesiscloud.Region(data.Region.ValueString())
	body.Description = pointer(data.Description.ValueString())

	response, err := r.client.CreateFloatingIPWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create floating_ip", err))
		return
	}

	floatingIPResponse := response.JSON201
	if floatingIPResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create floating_ip", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &floatingIPResponse.FloatingIp)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a floating_ip resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	floatingIPId := floatingIPResponse.FloatingIp.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling floatingIP", err))
			return
		}

		tflog.Trace(ctx, "polling a floatingIP resource")

		response, err := r.client.GetFloatingIPWithResponse(ctx, floatingIPId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling floatingIP", err))
			return
		}

		floatingIPResponse := response.JSON200
		if floatingIPResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling floatingIP", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := floatingIPResponse.FloatingIp.Status
		if status == "created" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &floatingIPResponse.FloatingIp)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling floatingIP", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *FloatingIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FloatingIPResourceModel

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

	floatingIPId := data.Id.ValueString()

	response, err := r.client.GetFloatingIPWithResponse(ctx, floatingIPId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read floating_ip", err))
		return
	}

	floatingIPResponse := response.JSON200
	if floatingIPResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read floating_ip", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &floatingIPResponse.FloatingIp)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a floating_ip resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FloatingIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FloatingIPResourceModel

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

	body := genesiscloud.UpdateFloatingIPJSONRequestBody{}

	body.Name = pointer(data.Name.ValueString())
	body.Description = data.Description.ValueStringPointer()

	floatingIPId := data.Id.ValueString()

	response, err := r.client.UpdateFloatingIPWithResponse(ctx, floatingIPId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update floating_ip", err))
		return
	}

	floatingIPResponse := response.JSON200
	if floatingIPResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update floating_ip", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &floatingIPResponse.FloatingIp)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a floating_ip resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FloatingIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FloatingIPResourceModel

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

	floatingIPId := data.Id.ValueString()

	response, err := r.client.DeleteFloatingIPWithResponse(ctx, floatingIPId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete floating_ip", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete floating_ip", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}
}

func (r *FloatingIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
