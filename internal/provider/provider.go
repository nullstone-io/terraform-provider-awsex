package provider

import (
	"context"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-awsex/internal/conns"
)

// Ensure AwsexProvider satisfies various provider interfaces.

var (
	_ provider.ProviderWithValidateConfig = &AwsexProvider{}
	_ provider.ProviderWithFunctions      = &AwsexProvider{}
)

// AwsexProvider defines the provider implementation.
type AwsexProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *AwsexProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsex"
	resp.Version = p.version
}

func (p *AwsexProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Extension to AWS Provider",
		MarkdownDescription: `
This is a Terraform provider to extend the functionality of the [AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs).

At time of creation, Hashicorp had few review resources dedicated to the provider.
As a result, there are over 400 open pull requests, many older than 4 years.

The purpose of this provider is to rapidly augment the official provider.
This can also be used to rapidly experiment with new resources.
`,
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Optional: true,
				Description: "The access key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"assume_role":                   assumeRoleSchema(),
			"assume_role_with_web_identity": assumeRoleWithWebIdentitySchema(),
			"custom_ca_bundle": schema.StringAttribute{
				Optional: true,
				Description: "File containing custom root and intermediate certificates. " +
					"Can also be configured using the `AWS_CA_BUNDLE` environment variable. " +
					"(Setting `ca_bundle` in the shared config file is not supported.)",
			},
			"http_proxy": schema.StringAttribute{
				Optional: true,
				Description: "URL of a proxy to use for HTTP requests when accessing the AWS API. " +
					"Can also be set using the `HTTP_PROXY` or `http_proxy` environment variables.",
			},
			"https_proxy": schema.StringAttribute{
				Optional: true,
				Description: "URL of a proxy to use for HTTPS requests when accessing the AWS API. " +
					"Can also be set using the `HTTPS_PROXY` or `https_proxy` environment variables.",
			},
			"insecure": schema.BoolAttribute{
				Optional: true,
				Description: "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, " +
					"default value is `false`",
			},
			"max_retries": schema.Int32Attribute{
				Optional: true,
				Description: "The maximum number of times an AWS API request is\n" +
					"being executed. If the API request still fails, an error is\n" +
					"thrown.",
			},
			"no_proxy": schema.StringAttribute{
				Optional: true,
				Description: "Comma-separated list of hosts that should not use HTTP or HTTPS proxies. " +
					"Can also be set using the `NO_PROXY` or `no_proxy` environment variables.",
			},
			"profile": schema.StringAttribute{
				Optional: true,
				Description: "The profile for API operations. If not set, the default profile\n" +
					"created with `aws configure` will be used.",
			},
			"region": schema.StringAttribute{
				Optional: true,
				Description: "The region where AWS operations will take place. Examples\n" +
					"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003,
			},
			"retry_mode": schema.StringAttribute{
				Optional: true,
				Description: "Specifies how retries are attempted. Valid values are `standard` and `adaptive`. " +
					"Can also be configured using the `AWS_RETRY_MODE` environment variable.",
			},
			"secret_key": schema.StringAttribute{
				Optional: true,
				Description: "The secret key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"shared_config_files": schema.ListAttribute{
				Optional:    true,
				Description: "List of paths to shared config files. If not set, defaults to [~/.aws/config].",
				ElementType: types.StringType,
			},
			"shared_credentials_files": schema.ListAttribute{
				Optional:    true,
				Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
				ElementType: types.StringType,
			},
			"token": schema.StringAttribute{
				Optional: true,
				Description: "session token. A session token is only required if you are\n" +
					"using temporary security credentials.",
			},
		},
	}
}

func (p *AwsexProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model AwsexProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Configuring Terraform AWS Provider")
	awsbaseConfig := model.GetAwsBaseConfig(p.version, req.TerraformVersion)
	ctx, cfg, basediags := awsbase.GetAwsConfig(ctx, &awsbaseConfig)
	for _, d := range basediags {
		switch int(d.Severity()) {
		case int(diag.SeverityError):
			resp.Diagnostics.AddError(d.Summary(), d.Detail())
		case int(diag.SeverityWarning):
			resp.Diagnostics.AddWarning(d.Summary(), d.Detail())
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client := &conns.Client{Config: cfg}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AwsexProvider) ValidateConfig(ctx context.Context, request provider.ValidateConfigRequest, response *provider.ValidateConfigResponse) {
}

func (p *AwsexProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudfrontDistributionInvalidationResource,
		NewCloudfrontDistributionInvalidationsResource,
	}
}

func (p *AwsexProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *AwsexProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AwsexProvider{
			version: version,
		}
	}
}
