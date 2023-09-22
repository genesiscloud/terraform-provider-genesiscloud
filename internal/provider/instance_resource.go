package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/defaultplanmodifier"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                     = &InstanceResource{}
	_ resource.ResourceWithConfigure        = &InstanceResource{}
	_ resource.ResourceWithImportState      = &InstanceResource{}
	_ resource.ResourceWithConfigValidators = &InstanceResource{}
)

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Instance resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this image was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"hostname": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The hostname of your instance. If not provided will be initially set to the `name` attribute.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the instance.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"image_id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The image of the instance.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"metadata": schema.SingleNestedAttribute{
				MarkdownDescription: "Option to provide metadata. Currently supported is `startup_script`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"startup_script": resourceenhancer.Attribute(ctx, schema.StringAttribute{
						MarkdownDescription: "A plain text bash script or \"cloud-config\" file that will be executed after the first instance boot. " +
							"It is limited to 64 KiB in size. You can use it to configure your instance, e.g. installing the NVIDIA GPU driver. " +
							"Learn more about [startup scripts and installing the GPU driver](https://support.genesiscloud.com/support/solutions/articles/47001122478).",
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					}),
				},
			},
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the instance.",
				Required:            true,
			}),
			"password": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The password to access the instance. " +
					"Your password must have upper and lower chars, digits and length between 8-72. " +
					"**Please Note**: Only one of `ssh_keys` or `password` can be provided. " +
					"Password is less secure - we recommend you use an SSH key-pair.",
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(16),
				},
			}),
			"placement_option": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The placement option identifier in which instances are physically located relative to each other within a zone.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					defaultplanmodifier.String("AUTO"),
				},
				Validators: []validator.String{},
			}),
			"private_ip": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The private IPv4 IP-Address (IPv4 address).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
					// TODO: Could be changed outside of terraform via stop+start?
				},
			}),
			"public_ip": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The public IPv4 IP-Address (IPv4 address).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
					// TODO: Could be changed outside of terraform via stop+start?
				},
			}),
			"public_ip_type": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: `When set to "static", the instance's public IP will not change between start and stop actions.`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					defaultplanmodifier.String(string(genesiscloud.InstancePublicIPTypeDynamic)),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(genesiscloud.AllInstancePublicIPTypes)...),
				},
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
			"security_group_ids": resourceenhancer.Attribute(ctx, schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The security groups of the instance. If not provided will be set to the default security group.",
				Optional:            true,
				Computed:            true, // might be changed outside of Terraform
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(), // if unset, expect no changes
				},
			}),
			"ssh_key_ids": resourceenhancer.Attribute(ctx, schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The ssh keys of the instance.",
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1), // Note: This is a internal limitation that should be lifted in the future
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The instance status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"type": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The instance type identifier. Learn more about instance types [here](https://developers.genesiscloud.com/instances#instance-types).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"updated_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this image was last updated in RFC 3339.",
				Computed:            true,
			}),
			"volume_ids": resourceenhancer.Attribute(ctx, schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The volumes of the instance.",
				Optional:            true,
				Computed:            true, // might be changed outside of Terraform
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(), // if unset, expect no changes
				},
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *InstanceResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("password"),
			path.MatchRoot("ssh_key_ids"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("metadata").AtName("startup_script"),
			// In the future add additional metadata options here
		),
	}
}

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel

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

	body := genesiscloud.CreateInstanceJSONRequestBody{}

	body.Name = data.Name.ValueString()
	body.Hostname = data.Hostname.ValueString()

	if body.Hostname == "" {
		body.Hostname = data.Name.ValueString()
	}

	body.Type = genesiscloud.InstanceType(data.Type.ValueString())
	body.Image = data.ImageId.ValueString()

	if data.Metadata != nil {
		body.Metadata = &struct {
			StartupScript *string                        `json:"startup_script,omitempty"`
			UserData      *genesiscloud.InstanceUserData `json:"user_data,omitempty"`
		}{
			StartupScript: pointer(data.Metadata.StartupScript.ValueString()),
		}
	}

	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		body.Password = pointer(data.Password.ValueString())
	}

	if !data.SecurityGroupIds.IsNull() && !data.SecurityGroupIds.IsUnknown() {
		var securityGroups []string
		data.SecurityGroupIds.ElementsAs(ctx, &securityGroups, false)
		body.SecurityGroups = &securityGroups
	}

	if !data.SshKeyIds.IsNull() && !data.SshKeyIds.IsUnknown() {
		var sshKeyIds []string
		data.SshKeyIds.ElementsAs(ctx, &sshKeyIds, false)
		body.SshKeys = &sshKeyIds
	}

	if !data.VolumeIds.IsNull() && !data.VolumeIds.IsUnknown() {
		var volumeIds []string
		data.VolumeIds.ElementsAs(ctx, &volumeIds, false)
		body.Volumes = &volumeIds
	}

	body.PublicIpType = pointer(genesiscloud.InstancePublicIPType(data.PublicIpType.ValueString()))
	body.Region = genesiscloud.Region(data.Region.ValueString())
	body.PlacementOption = pointer(data.PlacementOption.ValueString())

	response, err := r.client.CreateInstanceWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create instance", err))
		return
	}

	instanceResponse := response.JSON201
	if instanceResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create instance", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a instance resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceId := instanceResponse.Instance.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling instance", err))
			return
		}

		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling instance", err))
			return
		}

		instanceResponse := response.JSON200
		if instanceResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling instance", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := instanceResponse.Instance.Status
		if status == "active" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling instance", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel

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

	instanceId := data.Id.ValueString()

	response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read instance", err))
		return
	}

	instanceResponse := response.JSON200
	if instanceResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read instance", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a instance resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceResourceModel

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

	body := genesiscloud.UpdateInstanceJSONRequestBody{}

	body.Name = pointer(data.Name.ValueString())

	if !data.SecurityGroupIds.IsNull() && !data.SecurityGroupIds.IsUnknown() {
		var securityGroups []string
		data.SecurityGroupIds.ElementsAs(ctx, &securityGroups, false)
		body.SecurityGroups = &securityGroups
	}

	if !data.VolumeIds.IsNull() && !data.VolumeIds.IsUnknown() {
		var volumeIds []string
		data.VolumeIds.ElementsAs(ctx, &volumeIds, false)
		body.Volumes = &volumeIds
	}

	instanceId := data.Id.ValueString()

	response, err := r.client.UpdateInstanceWithResponse(ctx, instanceId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update instance", err))
		return
	}

	instanceResponse := response.JSON200
	if instanceResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update instance", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a instance resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel

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

	instanceId := data.Id.ValueString()

	response, err := r.client.DeleteInstanceWithResponse(ctx, instanceId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete instance", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete instance", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling instance", err))
			return
		}

		tflog.Trace(ctx, "polling a instance resource")

		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling instance", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
