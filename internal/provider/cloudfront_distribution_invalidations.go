package provider

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"strings"
	"time"
)

var _ resource.Resource = &CloudfrontDistributionInvalidationsResource{}

type CloudfrontDistributionInvalidationsResource struct {
	client *conns.Client
}

func NewCloudfrontDistributionInvalidationsResource() resource.Resource {
	return &CloudfrontDistributionInvalidationsResource{}
}

type CloudfrontDistributionInvalidationsModel struct {
	Id              types.String   `tfsdk:"id"`
	DistributionIds types.Set      `tfsdk:"distribution_ids"`
	Paths           types.Set      `tfsdk:"paths"`
	Statuses        types.Map      `tfsdk:"statuses"`
	Triggers        types.Map      `tfsdk:"triggers"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (r *CloudfrontDistributionInvalidationsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cloudfront_distribution_invalidations"
}

func (r *CloudfrontDistributionInvalidationsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "",

		Attributes: map[string]schema.Attribute{
			"distribution_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A list of Cloudfront Distribution IDs where an invalidation will be created.",
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
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
			"statuses": schema.MapAttribute{
				ElementType: types.StringType,
				Description: "The status of each invalidation indexed by the Cloudfront Distribution ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
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
				MarkdownDescription: "The ID of the invalidations.",
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

func (r *CloudfrontDistributionInvalidationsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CloudfrontDistributionInvalidationsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data CloudfrontDistributionInvalidationsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 30*time.Minute)
	response.Diagnostics.Append(diags...)
	distributionIds := make([]string, 0)
	response.Diagnostics.Append(data.DistributionIds.ElementsAs(ctx, &distributionIds, false)...)
	paths := make([]string, 0)
	response.Diagnostics.Append(data.Paths.ElementsAs(ctx, &paths, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	result, diags := cloudfront.CreateInvalidations(ctx, r.client, distributionIds, paths, createTimeout)
	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(r.setResult(ctx, &data, distributionIds, result)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.Triggers.IsUnknown() {
		data.Triggers = types.MapNull(types.StringType)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CloudfrontDistributionInvalidationsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data CloudfrontDistributionInvalidationsModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	distributionIds := make([]string, 0)
	response.Diagnostics.Append(data.DistributionIds.ElementsAs(ctx, &distributionIds, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	ids := map[string]string{}
	for i, id := range strings.Split(data.Id.ValueString(), ";") {
		if i >= len(distributionIds) {
			break
		}
		ids[distributionIds[i]] = id
	}

	results, diags := cloudfront.FindInvalidations(ctx, r.client, ids)
	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(r.setResult(ctx, &data, distributionIds, results)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *CloudfrontDistributionInvalidationsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
}

func (r *CloudfrontDistributionInvalidationsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

func (r *CloudfrontDistributionInvalidationsResource) setResult(ctx context.Context, model *CloudfrontDistributionInvalidationsModel,
	distributionIds []string, results map[string]*cftypes.Invalidation) diag.Diagnostics {

	ids := make([]string, 0)
	statuses := map[string]string{}
	for _, distributionId := range distributionIds {
		cur, ok := results[distributionId]
		if !ok {
			ids = append(ids, "")
			statuses[distributionId] = "unknown"
		}
		ids = append(ids, aws.ToString(cur.Id))
		statuses[distributionId] = aws.ToString(cur.Status)
	}
	model.Id = types.StringValue(strings.Join(ids, ";"))
	var diags diag.Diagnostics
	model.Statuses, diags = types.MapValueFrom(ctx, types.StringType, statuses)
	return diags
}
