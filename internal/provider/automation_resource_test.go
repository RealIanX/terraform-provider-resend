package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAutomationResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "resend" {}
resource "resend_automation" "test" {
  name   = "tf-acc-automation"
  status = "disabled"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("resend_automation.test", "id"),
					resource.TestCheckResourceAttr("resend_automation.test", "status", "disabled"),
				),
			},
			{
				Config: `
provider "resend" {}
resource "resend_automation" "test" {
  name   = "tf-acc-automation-updated"
  status = "disabled"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("resend_automation.test", "name", "tf-acc-automation-updated"),
				),
			},
		},
	})
}
