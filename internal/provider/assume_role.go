package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"time"
)

func assumeRoleSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				CustomType:  timetypes.GoDurationType{},
				Optional:    true,
				Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
				Validators:  []validator.String{validAssumeRoleDuration{}},
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Description: "A unique identifier that might be required when you assume a role in another account.",
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(2, 1224),
						stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@:\/\-]*`), ""),
					),
				},
			},
			"policy": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
			},
			"policy_arns": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						ArnValidator{},
					),
				},
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
				Validators: []validator.String{
					ArnValidator{},
				},
			},
			"session_name": schema.StringAttribute{
				Optional:    true,
				Description: "An identifier for the assumed role session.",
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(2, 64),
						stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
					),
				},
			},
			"source_identity": schema.StringAttribute{
				Optional:    true,
				Description: "Source identity specified by the principal assuming the role.",
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(2, 64),
						stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
					),
				},
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Assume role session tags.",
			},
			"transitive_tag_keys": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Assume role session tag keys to pass to any subsequent sessions.",
			},
		},
	}
}

func assumeRoleWithWebIdentitySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				CustomType:  timetypes.GoDurationType{},
				Optional:    true,
				Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
				Validators:  []validator.String{validAssumeRoleDuration{}},
			},
			"policy": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
			},
			"policy_arns": schema.SetAttribute{
				Optional:    true,
				Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(ArnValidator{}),
				},
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
				Validators:  []validator.String{ArnValidator{}},
			},
			"session_name": schema.StringAttribute{
				Optional:    true,
				Description: "An identifier for the assumed role session.",
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(2, 64),
						stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
					),
				},
			},
			"web_identity_token": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.All(
						stringvalidator.LengthBetween(4, 20000),
						stringvalidator.ConflictsWith(
							path.MatchRoot("assume_role_with_web_identity.0.web_identity_token"),
							path.MatchRoot("assume_role_with_web_identity.0.web_identity_token_file"),
						),
					),
				},
			},
			"web_identity_token_file": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("assume_role_with_web_identity.0.web_identity_token"),
						path.MatchRoot("assume_role_with_web_identity.0.web_identity_token_file"),
					),
				},
			},
		},
	}
}

var (
	_ validator.String = validAssumeRoleDuration{}
)

type validAssumeRoleDuration struct{}

func (v validAssumeRoleDuration) Description(ctx context.Context) string {
	return "string must be a valid duration"
}

func (v validAssumeRoleDuration) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validAssumeRoleDuration) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}
	value := request.ConfigValue.ValueString()
	if value == "" {
		return
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("%q cannot be parsed as a duration: %s", request.Path, err), "")
		return
	}

	if duration.Minutes() > 15 || duration.Hours() > 12 {
		response.Diagnostics.AddError(fmt.Sprintf("duration %q must be between 15 minutes (15m) and 12 hours (12h), inclusive", request.Path), "")
	}
}
