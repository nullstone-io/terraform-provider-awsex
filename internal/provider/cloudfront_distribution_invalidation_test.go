package provider

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	providerConfig = `
provider "awsex" {
}
`
)

func TestAccDistributionInvalidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	cdnId := testAccCreateCdn(t, "test", "www.example.com")
	config1 := providerConfig + fmt.Sprintf(`
resource "awsex_cloudfront_distribution_invalidation" "test" {
  distribution_id = %[1]q
  paths           = ["/*"]
}
`, cdnId)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("awsex_cloudfront_distribution_invalidation.test", "id"),
					resource.TestCheckResourceAttr("awsex_cloudfront_distribution_invalidation.test", "status", "Completed"),
				),
			},
		},
	})
}

func testAccCreateCdn(t *testing.T, name, domainName string) string {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	client := cloudfront.NewFromConfig(cfg)
	input := &cloudfront.CreateDistributionInput{
		DistributionConfig: &cftypes.DistributionConfig{
			Enabled: aws.Bool(false),
			DefaultCacheBehavior: &cftypes.DefaultCacheBehavior{
				AllowedMethods: &cftypes.AllowedMethods{
					Items:    []cftypes.Method{cftypes.MethodGet, cftypes.MethodHead},
					Quantity: aws.Int32(2),
					CachedMethods: &cftypes.CachedMethods{
						Items:    []cftypes.Method{cftypes.MethodGet, cftypes.MethodHead},
						Quantity: aws.Int32(2),
					},
				},
				TargetOriginId:       aws.String(name),
				ViewerProtocolPolicy: "allow-all",
				MinTTL:               aws.Int64(0),
				ForwardedValues: &cftypes.ForwardedValues{
					QueryString: aws.Bool(false),
					Cookies:     &cftypes.CookiePreference{Forward: cftypes.ItemSelectionAll},
				},
			},
			Restrictions: &cftypes.Restrictions{
				GeoRestriction: &cftypes.GeoRestriction{
					RestrictionType: "none",
					Quantity:        aws.Int32(0),
				},
			},
			Origins: &cftypes.Origins{
				Quantity: aws.Int32(1),
				Items: []cftypes.Origin{
					{
						DomainName: aws.String(domainName),
						Id:         aws.String(name),
						CustomOriginConfig: &cftypes.CustomOriginConfig{
							HTTPPort:             aws.Int32(80),
							HTTPSPort:            aws.Int32(443),
							OriginProtocolPolicy: "https-only",
							OriginSslProtocols: &cftypes.OriginSslProtocols{
								Quantity: aws.Int32(1),
								Items:    []cftypes.SslProtocol{cftypes.SslProtocolTLSv12},
							},
						},
					},
				},
			},
			CallerReference: aws.String(uuid.NewString()),
			Comment:         aws.String("Test Distribution for terraform-provider-awsex"),
			ViewerCertificate: &cftypes.ViewerCertificate{
				CloudFrontDefaultCertificate: aws.Bool(true),
			},
		},
	}
	out, err := client.CreateDistribution(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("create distribution had a nil result")
	}

	t.Cleanup(func() {
		client.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
			Id: out.Distribution.Id,
		})
	})

	return *out.Distribution.Id
}
