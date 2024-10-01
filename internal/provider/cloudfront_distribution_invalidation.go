package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-awsex/internal/conns"
	"github.com/hashicorp/terraform-provider-awsex/internal/provider/cloudfront"
	"regexp"
	"time"
)

var _ resource.Resource = &CloudfrontDistributionInvalidationResource{}

type CloudfrontDistributionInvalidationResource struct {
	client *conns.Client
}

func NewCloudfrontDistributionInvalidationResource() resource.Resource {
	return &CloudfrontDistributionInvalidationResource{}
}

type CloudfrontDistributionInvalidationModel struct {
	Id             types.String   `tfsdk:"id"`
	DistributionId types.String   `tfsdk:"distribution_id"`
	Paths          types.Set      `tfsdk:"paths"`
	Status         types.String   `tfsdk:"status"`
	Triggers       types.Map      `tfsdk:"triggers"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func (r *CloudfrontDistributionInvalidationResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cloudfront_distribution_invalidation"
}

func (r *CloudfrontDistributionInvalidationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "",

		Attributes: map[string]schema.Attribute{
			"distribution_id": schema.StringAttribute{
				MarkdownDescription: "The Cloudfront Distribution ID where an invalidation should be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paths": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A list of paths to invalidate. Each path *must* start with `/`.",
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexp.MustCompile(`^/`), "")),
				},
			},
			"status": schema.StringAttribute{
				Description: "The status of the invalidation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of triggers that, when changed, will force Terraform to create a new invalidation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invalidation.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *CloudfrontDistributionInvalidationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*conns.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CloudfrontDistributionInvalidationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data CloudfrontDistributionInvalidationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 30*time.Minute)
	response.Diagnostics.Append(diags...)
	paths := make([]string, 0)
	response.Diagnostics.Append(data.Paths.ElementsAs(ctx, &paths, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	inval, diags := cloudfront.CreateInvalidation(ctx, r.client, data.DistributionId.ValueString(), paths, createTimeout)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringPointerValue(inval.Id)
	data.Status = types.StringPointerValue(inval.Status)
	if data.Triggers.IsUnknown() {
		data.Triggers = types.MapNull(types.StringType)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CloudfrontDistributionInvalidationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data CloudfrontDistributionInvalidationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	inval, diags := cloudfront.FindInvalidation(ctx, r.client, data.DistributionId.ValueString(), data.Id.ValueString())
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if inval == nil {
		response.State.RemoveResource(ctx)
	} else {
		data.Id = types.StringPointerValue(inval.Id)
		data.Status = types.StringPointerValue(inval.Status)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CloudfrontDistributionInvalidationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *CloudfrontDistributionInvalidationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}
