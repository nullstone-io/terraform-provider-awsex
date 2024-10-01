package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDistributionInvalidations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	cdn1Id := testAccCreateCdn(t, "test1", "www.example1.com")
	cdn2Id := testAccCreateCdn(t, "test2", "www.example2.com")
	config1 := providerConfig + fmt.Sprintf(`
resource "awsex_cloudfront_distribution_invalidations" "test" {
  distribution_ids = [%[1]q, %[2]q]
  paths            = ["/*"]
}
`, cdn1Id, cdn2Id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("awsex_cloudfront_distribution_invalidations.test", "id"),
					resource.TestCheckResourceAttr("awsex_cloudfront_distribution_invalidations.test", "statuses.%", "2"),
					resource.TestCheckResourceAttr("awsex_cloudfront_distribution_invalidations.test", fmt.Sprintf("statuses.%s", cdn1Id), "Completed"),
					resource.TestCheckResourceAttr("awsex_cloudfront_distribution_invalidations.test", fmt.Sprintf("statuses.%s", cdn2Id), "Completed"),
				),
			},
		},
	})
}
