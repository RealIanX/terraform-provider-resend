package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEventResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "resend" {}
resource "resend_event" "test" {
  name = "tf-acc-event"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("resend_event.test", "id"),
					resource.TestCheckResourceAttr("resend_event.test", "name", "tf-acc-event"),
				),
			},
			{
				Config: `
provider "resend" {}
resource "resend_event" "test" {
  name   = "tf-acc-event-updated"
  schema = jsonencode({"email": "string", "user_id": "number"})
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("resend_event.test", "name", "tf-acc-event-updated"),
				),
			},
		},
	})
}
