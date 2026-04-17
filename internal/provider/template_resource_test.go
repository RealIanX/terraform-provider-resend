package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTemplateResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "resend" {}
resource "resend_template" "test" {
  name = "tf-acc-basic"
  html = "<p>Hello</p>"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("resend_template.test", "id"),
					resource.TestCheckResourceAttr("resend_template.test", "name", "tf-acc-basic"),
				),
			},
			{
				Config: `
provider "resend" {}
resource "resend_template" "test" {
  name    = "tf-acc-updated"
  html    = "<p>Updated</p>"
  subject = "Hello"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("resend_template.test", "name", "tf-acc-updated"),
					resource.TestCheckResourceAttr("resend_template.test", "subject", "Hello"),
				),
			},
		},
	})
}
