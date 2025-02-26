package provider

import (
	"context"
	"fmt"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/defaultplanmodifier"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &VolumeResource{}
	_ resource.ResourceWithConfigure   = &VolumeResource{}
	_ resource.ResourceWithImportState = &VolumeResource{}
)

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// VolumeResource defines the resource implementation.
type VolumeResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this volume was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description for the volume.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the volume.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the volume.",
				Required:            true,
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
			"size": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
				MarkdownDescription: "The storage size of this volume given in GiB.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The volume status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"type": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The storage type of the volume.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllVolumeTypes)...),
				},
			}),

			// Internal
			"retain_on_delete": resourceenhancer.Attribute(ctx, schema.BoolAttribute{
				MarkdownDescription: "Flag to retain the volume when the resource is deleted",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					defaultplanmodifier.Bool(false),
				},
			}),

			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

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

	body := genesiscloud.CreateVolumeJSONRequestBody{}

	body.Description = pointer(data.Description.ValueString())
	body.Name = data.Name.ValueString()
	body.Region = genesiscloud.Region(data.Region.ValueString())
	body.Size = int(data.Size.ValueInt64())
	body.Type = pointer(genesiscloud.VolumeType(data.Type.ValueString()))

	response, err := r.client.CreateVolumeWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create volume", err))
		return
	}

	volumeResponse := response.JSON201
	if volumeResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create volume", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &volumeResponse.Volume)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a volume resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	volumeId := volumeResponse.Volume.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling volume", err))
			return
		}

		tflog.Trace(ctx, "polling a volume resource")

		response, err := r.client.GetVolumeWithResponse(ctx, volumeId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling volume", err))
			return
		}

		volumeResponse := response.JSON200
		if volumeResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling volume", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := volumeResponse.Volume.Status
		if status == genesiscloud.VolumeStatusCreated || status == genesiscloud.VolumeStatusError {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &volumeResponse.Volume)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == genesiscloud.VolumeStatusError {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling volume", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VolumeResourceModel

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

	volumeId := data.Id.ValueString()

	response, err := r.client.GetVolumeWithResponse(ctx, volumeId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read volume", err))
		return
	}

	volumeResponse := response.JSON200
	if volumeResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read volume", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &volumeResponse.Volume)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a volume resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VolumeResourceModel

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

	body := genesiscloud.UpdateVolumeJSONRequestBody{}

	body.Name = pointer(data.Name.ValueString())
	body.Description = pointer(data.Description.ValueString())
	body.Size = pointer(int(data.Size.ValueInt64()))

	volumeId := data.Id.ValueString()

	response, err := r.client.UpdateVolumeWithResponse(ctx, volumeId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update volume", err))
		return
	}

	volumeResponse := response.JSON200
	if volumeResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update volume", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &volumeResponse.Volume)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a volume resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VolumeResourceModel

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

	volumeId := data.Id.ValueString()

	if data.RetainOnDelete.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Volume is retained",
			fmt.Sprintf("The volume resource with id %q was deleted from the state but the volume is retained.", volumeId),
		)
		return
	}

	response, err := r.client.DeleteVolumeWithResponse(ctx, volumeId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete volume", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete volume", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling volume", err))
			return
		}

		tflog.Trace(ctx, "polling a volume resource")

		response, err := r.client.GetVolumeWithResponse(ctx, volumeId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling volume", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
