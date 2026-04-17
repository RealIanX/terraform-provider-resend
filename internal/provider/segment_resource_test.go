package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSegmentResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "resend" {}
resource "resend_segment" "test" {
  name = "tf-acc-segment"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("resend_segment.test", "id"),
					resource.TestCheckResourceAttr("resend_segment.test", "name", "tf-acc-segment"),
				),
			},
		},
	})
}
