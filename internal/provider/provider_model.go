package provider

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
)

type AwsexProviderModel struct {
	// AccessKey
	// The access key for API operations. You can retrieve this from the 'Security & Credentials' section of the AWS console.
	AccessKey                 *string                              `tfsdk:"access_key"`
	AssumeRole                *AwsexAssumeRoleModel                `tfsdk:"assume_role"`
	AssumeRoleWithWebIdentity *AwsexAssumeRoleWithWebIdentityModel `tfsdk:"assume_role_with_web_identity"`
	// CustomCaBundle
	// File containing custom root and intermediate certificates.
	// Can also be configured using the `AWS_CA_BUNDLE` environment variable.
	// (Setting `ca_bundle` in the shared config file is not supported.)
	CustomCaBundle *string `tfsdk:"custom_ca_bundle"`
	// HttpProxy
	// URL of a proxy to use for HTTP requests when accessing the AWS API.
	// Can also be set using the `HTTP_PROXY` or `http_proxy` environment variables.
	HttpProxy *string `tfsdk:"http_proxy"`
	// HttpsProxy
	// URL of a proxy to use for HTTPS requests when accessing the AWS API.
	// Can also be set using the `HTTPS_PROXY` or `https_proxy` environment variables.
	HttpsProxy *string `tfsdk:"https_proxy"`
	// Insecure
	// Explicitly allow the provider to perform "insecure" SSL requests.
	// If omitted, default value is `false`
	Insecure *bool `tfsdk:"insecure"`
	// MaxRetries
	// The maximum number of times an AWS API request is being executed.
	// If the API request still fails, an error is thrown.
	MaxRetries *int `tfsdk:"max_retries"`
	// NoProxy
	// Comma-separated list of hosts that should not use HTTP or HTTPS proxies.
	// Can also be set using the `NO_PROXY` or `no_proxy` environment variables.
	NoProxy *string `tfsdk:"no_proxy"`
	// Profile
	// The profile for API operations. If not set, the default profile created with `aws configure` will be used.
	Profile *string `tfsdk:"profile"`
	// Region
	// The region where AWS operations will take place.
	// Examples are us-east-1, us-west-2, etc.
	Region *string `tfsdk:"region"`
	// RetryMode
	// Specifies how retries are attempted. Valid values are `standard` and `adaptive`.
	// Can also be configured using the `AWS_RETRY_MODE` environment variable.
	RetryMode *string `tfsdk:"retry_mode"`
	// SecretKey
	// The secret key for API operations. You can retrieve this from the 'Security & Credentials' section of the AWS console.
	SecretKey *string `tfsdk:"secret_key"`
	// SharedConfigFiles
	// List of paths to shared config files. If not set, defaults to [~/.aws/config].
	SharedConfigFiles []string `tfsdk:"shared_config_files"`
	// SharedCredentialsFiles
	// List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].
	SharedCredentialsFiles []string `tfsdk:"shared_credentials_files"`
	// Token
	// session token. A session token is only required if you are using temporary security credentials.
	Token *string `tfsdk:"token"`
}

func (m AwsexProviderModel) GetAwsBaseConfig(providerVersion, terraformVersion string) awsbase.Config {
	awsbaseConfig := awsbase.Config{
		AccessKey: unptr(m.AccessKey),
		APNInfo: &awsbase.APNInfo{
			PartnerName: "Nullstone",
			Products: []awsbase.UserAgentProduct{
				{Name: "Terraform", Version: terraformVersion, Comment: "+https://www.terraform.io"},
				{Name: "terraform-provider-awsex", Version: providerVersion, Comment: "+https://registry.terraform.io/providers/nullstone-io/awsex"},
			},
		},
		//Backoff:                       &v1CompatibleBackoff{maxRetryDelay: maxBackoff},
		CallerDocumentationURL: "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:             "Terraform AWS Provider",
		//EC2MetadataServiceEnableState: m.EC2MetadataServiceEnableState,
		Insecure: unptr(m.Insecure),
		//HTTPClient:    client.HTTPClient(ctx),
		HTTPProxy:     m.HttpProxy,
		HTTPSProxy:    m.HttpsProxy,
		HTTPProxyMode: awsbase.HTTPProxyModeLegacy,
		//Logger:                        logger,
		//MaxBackoff:                    maxBackoff,
		MaxRetries: unptr(m.MaxRetries),
		NoProxy:    unptr(m.NoProxy),
		Profile:    unptr(m.Profile),
		Region:     unptr(m.Region),
		RetryMode:  aws.RetryMode(unptr(m.RetryMode)),
		SecretKey:  unptr(m.SecretKey),
		Token:      unptr(m.Token),
	}
	m.AssumeRole.Configure(&awsbaseConfig)
	m.AssumeRoleWithWebIdentity.Configure(&awsbaseConfig)
	if len(m.SharedConfigFiles) != 0 {
		awsbaseConfig.SharedConfigFiles = m.SharedConfigFiles
	}
	if len(m.SharedCredentialsFiles) != 0 {
		awsbaseConfig.SharedCredentialsFiles = m.SharedCredentialsFiles
	}
	return awsbaseConfig
}

type AwsexAssumeRoleModel struct {
	Duration          timetypes.GoDuration `tfsdk:"duration"`
	ExternalId        string               `tfsdk:"external_id"`
	Policy            jsontypes.Normalized `tfsdk:"policy"`
	PolicyArns        []string             `tfsdk:"policy_arns"`
	RoleArn           string               `tfsdk:"role_arn"`
	SessionName       string               `tfsdk:"session_name"`
	SourceIdentity    string               `tfsdk:"source_identity"`
	Tags              map[string]string    `tfsdk:"tags"`
	TransitiveTagKeys []string             `tfsdk:"transitive_tag_keys"`
}

func (m *AwsexAssumeRoleModel) Configure(cfg *awsbase.Config) {
	if m == nil || m.RoleArn == "" {
		return
	}

	// Validation will catch errors from this conversion
	duration, _ := m.Duration.ValueGoDuration()

	cfg.AssumeRole = append(cfg.AssumeRole, awsbase.AssumeRole{
		RoleARN:           m.RoleArn,
		Duration:          duration,
		ExternalID:        m.ExternalId,
		Policy:            m.Policy.ValueString(),
		PolicyARNs:        m.PolicyArns,
		SessionName:       m.SessionName,
		SourceIdentity:    m.SourceIdentity,
		Tags:              m.Tags,
		TransitiveTagKeys: m.TransitiveTagKeys,
	})
}

type AwsexAssumeRoleWithWebIdentityModel struct {
	Duration             timetypes.GoDuration `tfsdk:"duration"`
	Policy               jsontypes.Normalized `tfsdk:"policy"`
	PolicyArns           []string             `tfsdk:"policy_arns"`
	RoleArn              string               `tfsdk:"role_arn"`
	SessionName          string               `tfsdk:"session_name"`
	WebIdentityToken     string               `tfsdk:"web_identity_token"`
	WebIdentityTokenFile string               `tfsdk:"web_identity_token_file"`
}

func (m *AwsexAssumeRoleWithWebIdentityModel) Configure(cfg *awsbase.Config) {
	if m == nil || m.RoleArn == "" {
		return
	}

	// Validation will catch errors from this conversion
	duration, _ := m.Duration.ValueGoDuration()

	cfg.AssumeRoleWithWebIdentity = &awsbase.AssumeRoleWithWebIdentity{
		RoleARN:              m.RoleArn,
		Duration:             duration,
		Policy:               m.Policy.ValueString(),
		PolicyARNs:           m.PolicyArns,
		SessionName:          m.SessionName,
		WebIdentityToken:     m.WebIdentityToken,
		WebIdentityTokenFile: m.WebIdentityTokenFile,
	}
}

func unptr[T any](val *T) T {
	var t T
	if val != nil {
		return *val
	}
	return t
}
