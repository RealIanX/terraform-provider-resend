resource "resend_template" "example" {
  name      = "Welcome Email"
  html      = "<p>Hello {{name}}, welcome!</p>"
  subject   = "Welcome!"
  from      = "hello@example.com"
  published = true
}
