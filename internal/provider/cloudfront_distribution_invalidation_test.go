package provider

import (
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

	config1 := providerConfig + `
resource "aws_cloudfront_distribution" "test" {
  enabled          = false
  retain_on_delete = false
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    forwarded_values {
      query_string = false
      cookies {
        forward = "all"
      }
    }
  }
  origin {
    domain_name = "www.example.com"
    origin_id   = "test"
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  viewer_certificate {
    cloudfront_default_certificate = true
  }
  lifecycle {
    ignore_changes = [web_acl_id]
  }
}

resource "awsex_cloudfront_distribution_invalidation" "test" {
  distribution_id = aws_cloudfront_distribution.test.id
  paths           = ["/*"]
}
`

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
