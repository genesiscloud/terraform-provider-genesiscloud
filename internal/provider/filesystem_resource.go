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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &FilesystemResource{}
	_ resource.ResourceWithConfigure   = &FilesystemResource{}
	_ resource.ResourceWithImportState = &FilesystemResource{}
)

func NewFilesystemResource() resource.Resource {
	return &FilesystemResource{}
}

// FilesystemResource defines the resource implementation.
type FilesystemResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *FilesystemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filesystem"
}

func (r *FilesystemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Filesystem resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this filesystem was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description for the filesystem.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "A unique identifier for each filesystem. This is automatically generated.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the filesystem.",
				Required:            true,
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The identifier for the region this filesystem exists in.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllRegions)...),
				},
			}),
			"size": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
				MarkdownDescription: "The storage size of this filesystem given in GiB.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The filesystem status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"mount_base_path": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The base path on the server under which the mount point can be accessed.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"mount_endpoint_range": resourceenhancer.Attribute(ctx, schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The start and end IP of the mount endpoint range. Expressed as a array with two entries.",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"type": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The storage type of the filesystem.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllFilesystemTypes)...),
				},
			}),

			// Internal
			"retain_on_delete": resourceenhancer.Attribute(ctx, schema.BoolAttribute{
				MarkdownDescription: "Flag to retain the filesystem when the resource is deleted",
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

func (r *FilesystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FilesystemResourceModel

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

	body := genesiscloud.CreateFilesystemJSONRequestBody{}

	body.Description = pointer(data.Description.ValueString())
	body.Name = data.Name.ValueString()
	body.Region = genesiscloud.Region(data.Region.ValueString())
	body.Size = int(data.Size.ValueInt64())
	body.Type = pointer(genesiscloud.FilesystemType(data.Type.ValueString()))

	response, err := r.client.CreateFilesystemWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create filesystem", err))
		return
	}

	filesystemResponse := response.JSON201
	if filesystemResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create filesystem", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &filesystemResponse.Filesystem)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a filesystem resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filesystemId := filesystemResponse.Filesystem.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling filesystem", err))
			return
		}

		tflog.Trace(ctx, "polling a filesystem resource")

		response, err := r.client.GetFilesystemWithResponse(ctx, filesystemId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling filesystem", err))
			return
		}

		filesystemResponse := response.JSON200
		if filesystemResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling filesystem", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := filesystemResponse.Filesystem.Status
		if status == "created" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &filesystemResponse.Filesystem)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling filesystem", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *FilesystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FilesystemResourceModel

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

	filesystemId := data.Id.ValueString()

	response, err := r.client.GetFilesystemWithResponse(ctx, filesystemId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read filesystem", err))
		return
	}

	filesystemResponse := response.JSON200
	if filesystemResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read filesystem", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &filesystemResponse.Filesystem)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a filesystem resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilesystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FilesystemResourceModel

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

	body := genesiscloud.UpdateFilesystemJSONRequestBody{}

	body.Name = pointer(data.Name.ValueString())
	body.Description = pointer(data.Description.ValueString())
	body.Size = pointer(int(data.Size.ValueInt64()))

	filesystemId := data.Id.ValueString()

	response, err := r.client.UpdateFilesystemWithResponse(ctx, filesystemId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update filesystem", err))
		return
	}

	filesystemResponse := response.JSON200
	if filesystemResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update filesystem", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &filesystemResponse.Filesystem)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a filesystem resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilesystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FilesystemResourceModel

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

	filesystemId := data.Id.ValueString()

	if data.RetainOnDelete.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Filesystem is retained",
			fmt.Sprintf("The filesystem resource with id %q was deleted from the state but the filesystem is retained.", filesystemId),
		)
		return
	}

	response, err := r.client.DeleteFilesystemWithResponse(ctx, filesystemId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete filesystem", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete filesystem", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling filesystem", err))
			return
		}

		tflog.Trace(ctx, "polling a filesystem resource")

		response, err := r.client.GetFilesystemWithResponse(ctx, filesystemId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling filesystem", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *FilesystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
