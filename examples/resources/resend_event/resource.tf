resource "resend_event" "example" {
  name   = "user.signup"
  schema = jsonencode({
    email = "string"
    plan  = "string"
  })
}
