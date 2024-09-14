package cloudfront

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-awsex/internal/conns"
)

var _ resource.Resource = &DistributionInvalidationResource{}

type DistributionInvalidationResource struct {
	client *conns.Client
}

func NewDistributionInvalidationResource() resource.Resource {
	return &DistributionInvalidationResource{}
}

type DistributionInvalidationModel struct {
	Id             types.String `tfsdk:"id"`
	DistributionId types.String `tfsdk:"distribution_id"`
	Paths          types.Set    `tfsdk:"paths"`
}

func (r *DistributionInvalidationResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_distribution_invalidation"
}

func (r *DistributionInvalidationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "",

		Attributes: map[string]schema.Attribute{
			"distribution_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"paths": schema.SetAttribute{
				MarkdownDescription: "",
				Required:            true,
				ElementType:         types.StringType,
				Validators:          []validator.Set{
					// TODO: Add path validator
				},
			},
			"triggers": schema.MapAttribute{
				MarkdownDescription: "",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DistributionInvalidationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DistributionInvalidationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DistributionInvalidationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.createInvalidation(ctx, data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *DistributionInvalidationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DistributionInvalidationModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *DistributionInvalidationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data DistributionInvalidationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.createInvalidation(ctx, data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *DistributionInvalidationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

func (r *DistributionInvalidationResource) createInvalidation(ctx context.Context, data DistributionInvalidationModel) (DistributionInvalidationModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	distributionId := data.DistributionId.ValueString()

	paths := make([]string, 0)
	diags = append(diags, data.Paths.ElementsAs(ctx, &paths, false)...)
	if diags.HasError() {
		return data, diags
	}

	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionId,
		InvalidationBatch: &types2.InvalidationBatch{
			CallerReference: aws.String("terraform-provider-awsex"),
			Paths: &types2.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}
	client := r.client.Cloudfront()
	out, err := client.CreateInvalidation(ctx, input)
	if err != nil {
		diags.AddError("Error creating AWS Cloudfront Invalidation", err.Error())
		return data, diags
	}
	if out != nil && out.Invalidation != nil && out.Invalidation.Id != nil {
		data.Id = types.StringValue(*out.Invalidation.Id)
		tflog.Trace(ctx, "Created Cloudfront Invalidation")
	} else {
		diags.AddWarning("Unable to create AWS Cloudfront Invalidation.", "AWS did not create an invalidation and gave no reason")
	}
	return data, diags
}
