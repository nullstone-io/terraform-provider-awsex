package provider

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"
)

func TestAccDistributionInvalidations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	cdn1Id := testAccCreateCdn(t)
	cdn2Id := testAccCreateCdn(t)
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
					testAccCheckResourceMapAttribute("awsex_cloudfront_distribution_invalidations.test", "statuses", map[string]string{
						cdn1Id: "Completed",
						cdn2Id: "Completed",
					}),
				),
			},
		},
	})
}

func testAccCheckResourceMapAttribute(resourceName, attributeName string, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// Retrieve the actual map attribute from the state
		attr, ok := rs.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("attribute not found: %s", attributeName)
		}

		// Parse the string-encoded map (assuming the attribute is stored as a JSON string or similar)
		var actualMap map[string]string
		err := json.Unmarshal([]byte(attr), &actualMap)
		if err != nil {
			return fmt.Errorf("failed to parse attribute %s: %s", attributeName, err)
		}

		// Compare the expected map with the actual map
		for k, v := range expected {
			if actualMap[k] != v {
				return fmt.Errorf("expected key %s to have value %s, got %s", k, v, actualMap[k])
			}
		}

		return nil
	}
}
