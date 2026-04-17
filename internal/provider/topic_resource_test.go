package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTopicResource_basic(t *testing.T) {
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "resend" {}
resource "resend_topic" "test" {
  name                 = "tf-acc-topic"
  default_subscription = "opt_in"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("resend_topic.test", "id"),
					resource.TestCheckResourceAttr("resend_topic.test", "default_subscription", "opt_in"),
					resource.TestCheckResourceAttr("resend_topic.test", "visibility", "private"),
				),
			},
			{
				Config: `
provider "resend" {}
resource "resend_topic" "test" {
  name                 = "tf-acc-topic-updated"
  default_subscription = "opt_in"
  visibility           = "public"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("resend_topic.test", "name", "tf-acc-topic-updated"),
					resource.TestCheckResourceAttr("resend_topic.test", "visibility", "public"),
				),
			},
		},
	})
}
