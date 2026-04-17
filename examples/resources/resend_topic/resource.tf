resource "resend_topic" "example" {
  name                 = "Newsletter"
  default_subscription = "opt_in"
  description          = "Monthly newsletter subscribers"
  visibility           = "public"
}
