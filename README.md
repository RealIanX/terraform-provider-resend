# terraform-provider-resend

Terraform provider for [Resend](https://resend.com) — manage email templates, segments, topics, automations, and events as infrastructure.

## Requirements

- Terraform >= 1.0
- Go >= 1.22 (for development)
- A [Resend API key](https://resend.com/api-keys)

## Usage

```hcl
terraform {
  required_providers {
    resend = {
      source  = "realianx/resend"
      version = "~> 0.1"
    }
  }
}

provider "resend" {
  # api_key = "re_xxx"  # or set RESEND_API_KEY env var
}
```

## Resources

### `resend_template`

```hcl
resource "resend_template" "welcome" {
  name      = "Welcome Email"
  html      = "<p>Hello {{name}}, welcome!</p>"
  subject   = "Welcome!"
  from      = "hello@example.com"
  published = true
}
```

### `resend_segment`

```hcl
resource "resend_segment" "active_users" {
  name = "Active Users"
}
```

Note: The Resend API has no update endpoint for segments. Changing the `name` will destroy and recreate the segment.

### `resend_topic`

```hcl
resource "resend_topic" "newsletter" {
  name                 = "Newsletter"
  default_subscription = "opt_in"  # or "opt_out"
  visibility           = "public"  # or "private" (default)
}
```

Note: `default_subscription` is immutable after creation.

### `resend_automation`

```hcl
resource "resend_automation" "onboarding" {
  name   = "Onboarding Flow"
  status = "enabled"  # or "disabled" (default)

  steps = jsonencode([
    # See https://resend.com/docs/api-reference/automations for step schema
  ])
  connections = jsonencode([])
}
```

### `resend_event`

```hcl
resource "resend_event" "user_signup" {
  name   = "user.signup"
  schema = jsonencode({
    email = "string"
    plan  = "string"
  })
}
```

Note: Event names cannot start with `resend:`.

## Development

```bash
# Build
go build ./...

# Unit tests
go test ./resend/... -v

# Acceptance tests (requires real API key)
RESEND_API_KEY=re_xxx TF_ACC=1 make testacc

# Generate docs
make docs
```

## License

MIT
