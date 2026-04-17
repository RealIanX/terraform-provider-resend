# terraform-provider-resend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and publish a public Terraform provider for the Resend API supporting templates, segments, topics, automations, and events.

**Architecture:** Mirror terraform-provider-dokploy — `resend/` HTTP client package + `internal/provider/` resource files. Fixed base URL `https://api.resend.com`, Bearer auth, one file per API domain.

**Tech Stack:** Go 1.25, `hashicorp/terraform-plugin-framework v1.19.0`, goreleaser, GitHub Actions.

---

## File Map

| File | Responsibility |
|------|----------------|
| `main.go` | Provider entry point |
| `go.mod` | Module `github.com/realianx/terraform-provider-resend` |
| `GNUmakefile` | build / test / docs |
| `.goreleaser.yml` | Multi-platform release + GPG |
| `.github/workflows/test.yml` | CI |
| `.github/workflows/release.yml` | Release on `v*` tag |
| `resend/client.go` | HTTP client, Bearer auth, `HTTPError` |
| `resend/client_test.go` | Client unit tests (httptest) |
| `resend/template.go` | Template CRUD + publish |
| `resend/segment.go` | Segment create/get/delete |
| `resend/topic.go` | Topic CRUD |
| `resend/automation.go` | Automation CRUD + stop |
| `resend/event.go` | Event CRUD |
| `internal/provider/provider.go` | Provider config + registration |
| `internal/provider/provider_test.go` | Acc test setup + `testAccPreCheck` |
| `internal/provider/template_resource.go` | `resend_template` |
| `internal/provider/template_resource_test.go` | Template acc tests |
| `internal/provider/segment_resource.go` | `resend_segment` |
| `internal/provider/segment_resource_test.go` | Segment acc tests |
| `internal/provider/topic_resource.go` | `resend_topic` |
| `internal/provider/topic_resource_test.go` | Topic acc tests |
| `internal/provider/automation_resource.go` | `resend_automation` |
| `internal/provider/automation_resource_test.go` | Automation acc tests |
| `internal/provider/event_resource.go` | `resend_event` |
| `internal/provider/event_resource_test.go` | Event acc tests |
| `examples/provider/main.tf` | Provider usage example |
| `CLAUDE.md` | Dev instructions |
| `README.md` | Public documentation |

---

## Task 1: Go module scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `GNUmakefile`

- [ ] **Step 1: Create `go.mod`**

```
module github.com/realianx/terraform-provider-resend

go 1.25.0

require (
	github.com/hashicorp/terraform-plugin-framework v1.19.0
	github.com/hashicorp/terraform-plugin-go v0.31.0
	github.com/hashicorp/terraform-plugin-testing v1.15.0
)
```

Run `go mod tidy` to populate `go.sum`.

- [ ] **Step 2: Create `main.go`**

```go
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/realianx/terraform-provider-resend/internal/provider"
)

var version string = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run with debugger support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/realianx/resend",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
```

- [ ] **Step 3: Create `GNUmakefile`**

```makefile
default: build

.PHONY: build
build:
	go build -o terraform-provider-resend .

.PHONY: test
test:
	go test ./... -v

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider/... -v -count=1 -timeout 120m

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

.PHONY: fmt
fmt:
	gofmt -s -w .
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./...
```

Expected: no output (compiles cleanly once `internal/provider` exists).

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go GNUmakefile
git commit -m "feat: scaffold Go module"
```

---

## Task 2: HTTP client

**Files:**
- Create: `resend/client.go`
- Create: `resend/client_test.go`

- [ ] **Step 1: Write failing test**

Create `resend/client_test.go`:

```go
package resend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/realianx/terraform-provider-resend/resend"
)

func newTestClient(srv *httptest.Server) *resend.Client {
	c := resend.NewClient("test-key")
	c.BaseURL = srv.URL
	c.HTTPClient = srv.Client()
	return c
}

func TestClientDo_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header, got: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "abc"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	var out struct{ ID string `json:"id"` }
	if err := c.Do(context.Background(), http.MethodGet, "/test", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "abc" {
		t.Errorf("expected abc, got %s", out.ID)
	}
}

func TestClientDo_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	err := c.Do(context.Background(), http.MethodGet, "/missing", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := resend.AsHTTPError(err)
	if !ok || httpErr.StatusCode != 404 {
		t.Errorf("expected 404 HTTPError, got: %v", err)
	}
}
```

- [ ] **Step 2: Run test — expect compile failure**

```bash
go test ./resend/... 2>&1 | head -20
```

Expected: `cannot find package "github.com/realianx/terraform-provider-resend/resend"`

- [ ] **Step 3: Create `resend/client.go`**

```go
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const DefaultBaseURL = "https://api.resend.com"

type HTTPError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("resend: %s %s: status %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

func AsHTTPError(err error) (*HTTPError, bool) {
	var e *HTTPError
	return e, errors.As(err, &e)
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("resend: encode body: %w", err)
		}
		bodyReader = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("resend: new request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", "terraform-provider-resend")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &HTTPError{
			Method:     method,
			Path:       path,
			StatusCode: resp.StatusCode,
			Body:       string(bytes.TrimSpace(errBody)),
		}
	}

	if out != nil && resp.ContentLength != 0 {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("resend: decode response: %w", err)
		}
	}

	return nil
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./resend/... -v -run TestClient
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add resend/client.go resend/client_test.go
git commit -m "feat: add resend HTTP client"
```

---

## Task 3: Provider skeleton

**Files:**
- Create: `internal/provider/provider.go`
- Create: `internal/provider/provider_test.go`

- [ ] **Step 1: Create `internal/provider/provider_test.go`**

```go
package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/realianx/terraform-provider-resend/internal/provider"
)

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("RESEND_API_KEY") == "" {
		t.Skip("RESEND_API_KEY must be set for acceptance tests")
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"resend": providerserver.NewProtocol6WithError(provider.New("test")()),
}
```

- [ ] **Step 2: Create `internal/provider/provider.go`**

```go
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ provider.Provider = &ResendProvider{}

type ResendProvider struct {
	version string
}

type ResendProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ResendProvider{version: version}
	}
}

func (p *ResendProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "resend"
	resp.Version = p.version
}

func (p *ResendProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Resend email API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Resend API key. Can also be set via RESEND_API_KEY environment variable.",
			},
		},
	}
}

func (p *ResendProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ResendProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Resend API Key",
			"Set the api_key attribute or RESEND_API_KEY environment variable.",
		)
		return
	}

	client := resend.NewClient(apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ResendProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ResendTemplateResource,
		ResendSegmentResource,
		ResendTopicResource,
		ResendAutomationResource,
		ResendEventResource,
	}
}

func (p *ResendProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
```

- [ ] **Step 3: Create stub files so the package compiles**

Create `internal/provider/template_resource.go` with just the constructor:

```go
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ResendTemplateResource() resource.Resource   { return nil }
func ResendSegmentResource() resource.Resource    { return nil }
func ResendTopicResource() resource.Resource      { return nil }
func ResendAutomationResource() resource.Resource { return nil }
func ResendEventResource() resource.Resource      { return nil }
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/provider.go internal/provider/provider_test.go internal/provider/template_resource.go
git commit -m "feat: add provider skeleton"
```

---

## Task 4: Template client

**Files:**
- Create: `resend/template.go`
- Create: `resend/template_test.go`

- [ ] **Step 1: Write failing tests**

Create `resend/template_test.go`:

```go
package resend_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/realianx/terraform-provider-resend/resend"
)

func TestCreateTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/templates" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "tpl-1", "object": "template"})
	}))
	defer srv.Close()
	c := newTestClient(srv)

	id, err := c.CreateTemplate(context.Background(), resend.CreateTemplateRequest{
		Name: "Welcome", HTML: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "tpl-1" {
		t.Errorf("expected tpl-1, got %s", id)
	}
}

func TestGetTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/templates/tpl-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(resend.Template{ID: "tpl-1", Name: "Welcome", HTML: "<p>Hi</p>"})
	}))
	defer srv.Close()
	c := newTestClient(srv)

	tpl, err := c.GetTemplate(context.Background(), "tpl-1")
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "Welcome" {
		t.Errorf("expected Welcome, got %s", tpl.Name)
	}
}

func TestDeleteTemplate(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/templates/tpl-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newTestClient(srv)

	if err := c.DeleteTemplate(context.Background(), "tpl-1"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("delete not called")
	}
}

func TestPublishTemplate(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/templates/tpl-1/publish" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := newTestClient(srv)

	if err := c.PublishTemplate(context.Background(), "tpl-1"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("publish not called")
	}
}
```

- [ ] **Step 2: Run — expect compile error**

```bash
go test ./resend/... -run TestCreateTemplate 2>&1 | head -5
```

Expected: `undefined: resend.CreateTemplateRequest`

- [ ] **Step 3: Create `resend/template.go`**

```go
package resend

import (
	"context"
	"net/http"
)

type Template struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Alias   string `json:"alias"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	ReplyTo string `json:"reply_to"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

type CreateTemplateRequest struct {
	Name    string `json:"name"`
	HTML    string `json:"html"`
	Alias   string `json:"alias,omitempty"`
	From    string `json:"from,omitempty"`
	Subject string `json:"subject,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
	Text    string `json:"text,omitempty"`
}

type UpdateTemplateRequest struct {
	Name    string `json:"name,omitempty"`
	HTML    string `json:"html,omitempty"`
	Alias   string `json:"alias,omitempty"`
	From    string `json:"from,omitempty"`
	Subject string `json:"subject,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
	Text    string `json:"text,omitempty"`
}

func (c *Client) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/templates", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetTemplate(ctx context.Context, id string) (*Template, error) {
	var t Template
	if err := c.Do(ctx, http.MethodGet, "/templates/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) UpdateTemplate(ctx context.Context, id string, req UpdateTemplateRequest) error {
	return c.Do(ctx, http.MethodPatch, "/templates/"+id, req, nil)
}

func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/templates/"+id, nil, nil)
}

func (c *Client) PublishTemplate(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodPost, "/templates/"+id+"/publish", nil, nil)
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
go test ./resend/... -v -run "TestCreateTemplate|TestGetTemplate|TestDeleteTemplate|TestPublishTemplate"
```

Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add resend/template.go resend/template_test.go
git commit -m "feat: add template client"
```

---

## Task 5: Template resource

**Files:**
- Modify: `internal/provider/template_resource.go` (replace stub)
- Create: `internal/provider/template_resource_test.go`

- [ ] **Step 1: Write acceptance test**

Create `internal/provider/template_resource_test.go`:

```go
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
```

- [ ] **Step 2: Implement `internal/provider/template_resource.go`**

Replace the file entirely:

```go
package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ resource.Resource = &TemplateResource{}
var _ resource.ResourceWithImportState = &TemplateResource{}
var _ resource.ResourceWithConfigure = &TemplateResource{}

type TemplateResource struct{ client *resend.Client }

type TemplateResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	HTML      types.String `tfsdk:"html"`
	Alias     types.String `tfsdk:"alias"`
	From      types.String `tfsdk:"from"`
	Subject   types.String `tfsdk:"subject"`
	ReplyTo   types.String `tfsdk:"reply_to"`
	Text      types.String `tfsdk:"text"`
	Published types.Bool   `tfsdk:"published"`
}

func ResendTemplateResource() resource.Resource { return &TemplateResource{} }

func (r *TemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *TemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computed := func() []planmodifier.String {
		return []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	}
	resp.Schema = schema.Schema{
		Description: "Manages a Resend email template.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, Description: "Template UUID.", PlanModifiers: computed()},
			"name":    schema.StringAttribute{Required: true, Description: "Template name."},
			"html":    schema.StringAttribute{Required: true, Description: "HTML body."},
			"alias":   schema.StringAttribute{Optional: true, Computed: true, Description: "Template alias.", PlanModifiers: computed()},
			"from":    schema.StringAttribute{Optional: true, Computed: true, Description: "Default sender.", PlanModifiers: computed()},
			"subject": schema.StringAttribute{Optional: true, Computed: true, Description: "Default subject.", PlanModifiers: computed()},
			"reply_to": schema.StringAttribute{Optional: true, Computed: true, Description: "Default reply-to.", PlanModifiers: computed()},
			"text":    schema.StringAttribute{Optional: true, Computed: true, Description: "Plain-text body.", PlanModifiers: computed()},
			"published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Set true to publish. Resend has no unpublish endpoint; this only triggers a publish call.",
			},
		},
	}
}

func (r *TemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*resend.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *TemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateTemplate(ctx, resend.CreateTemplateRequest{
		Name:    plan.Name.ValueString(),
		HTML:    plan.HTML.ValueString(),
		Alias:   plan.Alias.ValueString(),
		From:    plan.From.ValueString(),
		Subject: plan.Subject.ValueString(),
		ReplyTo: plan.ReplyTo.ValueString(),
		Text:    plan.Text.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Template", err.Error())
		return
	}
	plan.ID = types.StringValue(id)

	if plan.Published.ValueBool() {
		if err := r.client.PublishTemplate(ctx, id); err != nil {
			resp.Diagnostics.AddError("Error Publishing Template", err.Error())
			return
		}
	}

	t, err := r.client.GetTemplate(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Template", err.Error())
		return
	}
	applyTemplateToModel(&plan, t)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.client.GetTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Template", err.Error())
		return
	}
	applyTemplateToModel(&state, t)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if err := r.client.UpdateTemplate(ctx, id, resend.UpdateTemplateRequest{
		Name:    plan.Name.ValueString(),
		HTML:    plan.HTML.ValueString(),
		Alias:   plan.Alias.ValueString(),
		From:    plan.From.ValueString(),
		Subject: plan.Subject.ValueString(),
		ReplyTo: plan.ReplyTo.ValueString(),
		Text:    plan.Text.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error Updating Template", err.Error())
		return
	}

	if plan.Published.ValueBool() {
		if err := r.client.PublishTemplate(ctx, id); err != nil {
			resp.Diagnostics.AddError("Error Publishing Template", err.Error())
			return
		}
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTemplate(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Template", err.Error())
	}
}

func (r *TemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func applyTemplateToModel(m *TemplateResourceModel, t *resend.Template) {
	m.Name = types.StringValue(t.Name)
	m.HTML = types.StringValue(t.HTML)
	m.Alias = types.StringValue(t.Alias)
	m.From = types.StringValue(t.From)
	m.Subject = types.StringValue(t.Subject)
	m.ReplyTo = types.StringValue(t.ReplyTo)
	m.Text = types.StringValue(t.Text)
}

func isNotFound(err error) bool {
	var e *resend.HTTPError
	return errors.As(err, &e) && e.StatusCode == 404
}
```

- [ ] **Step 3: Verify build**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Run unit tests**

```bash
go test ./resend/... -v
```

Expected: all `PASS`

- [ ] **Step 5: (Optional) Run acceptance test with real key**

```bash
RESEND_API_KEY=re_xxx TF_ACC=1 go test ./internal/provider/... -v -run TestAccTemplateResource -timeout 10m
```

- [ ] **Step 6: Commit**

```bash
git add internal/provider/template_resource.go internal/provider/template_resource_test.go
git commit -m "feat: add resend_template resource"
```

---

## Task 6: Segment client + resource

**Files:**
- Create: `resend/segment.go`
- Modify: `internal/provider/template_resource.go` → split stub into `internal/provider/segment_resource.go`
- Create: `internal/provider/segment_resource_test.go`

- [ ] **Step 1: Create `resend/segment.go`**

```go
package resend

import (
	"context"
	"net/http"
)

type Segment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) CreateSegment(ctx context.Context, name string) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/segments", map[string]string{"name": name}, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetSegment(ctx context.Context, id string) (*Segment, error) {
	var s Segment
	if err := c.Do(ctx, http.MethodGet, "/segments/"+id, nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) DeleteSegment(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/segments/"+id, nil, nil)
}
```

- [ ] **Step 2: Create `internal/provider/segment_resource.go`**

Remove the stub `ResendSegmentResource` from `template_resource.go` (delete the line), then create:

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ resource.Resource = &SegmentResource{}
var _ resource.ResourceWithImportState = &SegmentResource{}
var _ resource.ResourceWithConfigure = &SegmentResource{}

type SegmentResource struct{ client *resend.Client }

type SegmentResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func ResendSegmentResource() resource.Resource { return &SegmentResource{} }

func (r *SegmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (r *SegmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend segment. Name changes force replacement (no update endpoint).",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, Description: "Segment UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, Description: "Segment name.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		},
	}
}

func (r *SegmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*resend.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SegmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SegmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateSegment(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Segment", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SegmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SegmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSegment(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Segment", err.Error())
		return
	}
	state.Name = types.StringValue(s.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SegmentResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No update endpoint — schema uses RequiresReplace on name, so this is never called.
}

func (r *SegmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SegmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSegment(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Segment", err.Error())
	}
}

func (r *SegmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

- [ ] **Step 3: Create `internal/provider/segment_resource_test.go`**

```go
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
```

- [ ] **Step 4: Verify build and unit tests**

```bash
go build ./... && go test ./resend/... -v
```

Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add resend/segment.go internal/provider/segment_resource.go internal/provider/segment_resource_test.go internal/provider/template_resource.go
git commit -m "feat: add resend_segment resource"
```

---

## Task 7: Topic client + resource

**Files:**
- Create: `resend/topic.go`
- Create: `internal/provider/topic_resource.go`
- Create: `internal/provider/topic_resource_test.go`

- [ ] **Step 1: Create `resend/topic.go`**

```go
package resend

import (
	"context"
	"net/http"
)

type Topic struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	DefaultSubscription string `json:"default_subscription"`
	Description         string `json:"description"`
	Visibility          string `json:"visibility"`
}

type CreateTopicRequest struct {
	Name                string `json:"name"`
	DefaultSubscription string `json:"default_subscription"`
	Description         string `json:"description,omitempty"`
	Visibility          string `json:"visibility,omitempty"`
}

type UpdateTopicRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
}

func (c *Client) CreateTopic(ctx context.Context, req CreateTopicRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/topics", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetTopic(ctx context.Context, id string) (*Topic, error) {
	var t Topic
	if err := c.Do(ctx, http.MethodGet, "/topics/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) UpdateTopic(ctx context.Context, id string, req UpdateTopicRequest) error {
	return c.Do(ctx, http.MethodPatch, "/topics/"+id, req, nil)
}

func (c *Client) DeleteTopic(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/topics/"+id, nil, nil)
}
```

- [ ] **Step 2: Create `internal/provider/topic_resource.go`**

Remove stub `ResendTopicResource` from the stub file, then create:

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ resource.Resource = &TopicResource{}
var _ resource.ResourceWithImportState = &TopicResource{}
var _ resource.ResourceWithConfigure = &TopicResource{}

type TopicResource struct{ client *resend.Client }

type TopicResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	DefaultSubscription types.String `tfsdk:"default_subscription"`
	Description         types.String `tfsdk:"description"`
	Visibility          types.String `tfsdk:"visibility"`
}

func ResendTopicResource() resource.Resource { return &TopicResource{} }

func (r *TopicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topic"
}

func (r *TopicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend topic.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, Description: "Topic UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, Description: "Topic name. Max 50 characters."},
			"default_subscription": schema.StringAttribute{
				Required:      true,
				Description:   "opt_in or opt_out. Immutable after creation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{Optional: true, Computed: true, Description: "Topic description. Max 200 characters.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"visibility":  schema.StringAttribute{Optional: true, Computed: true, Description: "public or private.", Default: stringdefault.StaticString("private")},
		},
	}
}

func (r *TopicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*resend.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *TopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateTopic(ctx, resend.CreateTopicRequest{
		Name:                plan.Name.ValueString(),
		DefaultSubscription: plan.DefaultSubscription.ValueString(),
		Description:         plan.Description.ValueString(),
		Visibility:          plan.Visibility.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Topic", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.client.GetTopic(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Topic", err.Error())
		return
	}
	state.Name = types.StringValue(t.Name)
	state.DefaultSubscription = types.StringValue(t.DefaultSubscription)
	state.Description = types.StringValue(t.Description)
	state.Visibility = types.StringValue(t.Visibility)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state TopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateTopic(ctx, state.ID.ValueString(), resend.UpdateTopicRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Visibility:  plan.Visibility.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error Updating Topic", err.Error())
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTopic(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Topic", err.Error())
	}
}

func (r *TopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

- [ ] **Step 3: Create `internal/provider/topic_resource_test.go`**

```go
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
```

- [ ] **Step 4: Build and test**

```bash
go build ./... && go test ./resend/... -v
```

- [ ] **Step 5: Commit**

```bash
git add resend/topic.go internal/provider/topic_resource.go internal/provider/topic_resource_test.go
git commit -m "feat: add resend_topic resource"
```

---

## Task 8: Automation client + resource

**Files:**
- Create: `resend/automation.go`
- Create: `internal/provider/automation_resource.go`
- Create: `internal/provider/automation_resource_test.go`

- [ ] **Step 1: Create `resend/automation.go`**

```go
package resend

import (
	"context"
	"encoding/json"
	"net/http"
)

type Automation struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	Steps       json.RawMessage `json:"steps"`
	Connections json.RawMessage `json:"connections"`
}

type CreateAutomationRequest struct {
	Name        string          `json:"name"`
	Status      string          `json:"status,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
	Connections json.RawMessage `json:"connections,omitempty"`
}

type UpdateAutomationRequest struct {
	Name        string          `json:"name,omitempty"`
	Status      string          `json:"status,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
	Connections json.RawMessage `json:"connections,omitempty"`
}

func (c *Client) CreateAutomation(ctx context.Context, req CreateAutomationRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/automations", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetAutomation(ctx context.Context, id string) (*Automation, error) {
	var a Automation
	if err := c.Do(ctx, http.MethodGet, "/automations/"+id, nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) UpdateAutomation(ctx context.Context, id string, req UpdateAutomationRequest) error {
	return c.Do(ctx, http.MethodPatch, "/automations/"+id, req, nil)
}

func (c *Client) DeleteAutomation(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/automations/"+id, nil, nil)
}
```

- [ ] **Step 2: Create `internal/provider/automation_resource.go`**

Remove stub `ResendAutomationResource`, then create:

```go
package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ resource.Resource = &AutomationResource{}
var _ resource.ResourceWithImportState = &AutomationResource{}
var _ resource.ResourceWithConfigure = &AutomationResource{}

type AutomationResource struct{ client *resend.Client }

type AutomationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Status      types.String `tfsdk:"status"`
	Steps       types.String `tfsdk:"steps"`
	Connections types.String `tfsdk:"connections"`
}

func ResendAutomationResource() resource.Resource { return &AutomationResource{} }

func (r *AutomationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

func (r *AutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend automation. steps and connections are JSON strings (use jsonencode()).",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, Description: "Automation UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":   schema.StringAttribute{Required: true, Description: "Automation name."},
			"status": schema.StringAttribute{Optional: true, Computed: true, Description: "enabled or disabled.", Default: stringdefault.StaticString("disabled")},
			"steps": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON array of automation steps. Must be provided with connections.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connections": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON array of step connections. Must be provided with steps.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *AutomationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*resend.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func rawJSON(s types.String) json.RawMessage {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return nil
	}
	return json.RawMessage(s.ValueString())
}

func (r *AutomationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AutomationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateAutomation(ctx, resend.CreateAutomationRequest{
		Name:        plan.Name.ValueString(),
		Status:      plan.Status.ValueString(),
		Steps:       rawJSON(plan.Steps),
		Connections: rawJSON(plan.Connections),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Automation", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AutomationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AutomationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	a, err := r.client.GetAutomation(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Automation", err.Error())
		return
	}
	state.Name = types.StringValue(a.Name)
	state.Status = types.StringValue(a.Status)
	if len(a.Steps) > 0 {
		state.Steps = types.StringValue(string(a.Steps))
	}
	if len(a.Connections) > 0 {
		state.Connections = types.StringValue(string(a.Connections))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AutomationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AutomationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state AutomationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateAutomation(ctx, state.ID.ValueString(), resend.UpdateAutomationRequest{
		Name:        plan.Name.ValueString(),
		Status:      plan.Status.ValueString(),
		Steps:       rawJSON(plan.Steps),
		Connections: rawJSON(plan.Connections),
	}); err != nil {
		resp.Diagnostics.AddError("Error Updating Automation", err.Error())
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AutomationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AutomationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAutomation(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Automation", err.Error())
	}
}

func (r *AutomationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

- [ ] **Step 3: Create `internal/provider/automation_resource_test.go`**

```go
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
```

- [ ] **Step 4: Build**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add resend/automation.go internal/provider/automation_resource.go internal/provider/automation_resource_test.go
git commit -m "feat: add resend_automation resource"
```

---

## Task 9: Event client + resource

**Files:**
- Create: `resend/event.go`
- Create: `internal/provider/event_resource.go`
- Create: `internal/provider/event_resource_test.go`

- [ ] **Step 1: Create `resend/event.go`**

```go
package resend

import (
	"context"
	"net/http"
)

type Event struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type CreateEventRequest struct {
	Name   string `json:"name"`
	Schema any    `json:"schema,omitempty"`
}

type UpdateEventRequest struct {
	Name   string `json:"name,omitempty"`
	Schema any    `json:"schema,omitempty"`
}

func (c *Client) CreateEvent(ctx context.Context, req CreateEventRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/events", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetEvent(ctx context.Context, id string) (*Event, error) {
	var e Event
	if err := c.Do(ctx, http.MethodGet, "/events/"+id, nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (c *Client) UpdateEvent(ctx context.Context, id string, req UpdateEventRequest) error {
	return c.Do(ctx, http.MethodPatch, "/events/"+id, req, nil)
}

func (c *Client) DeleteEvent(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/events/"+id, nil, nil)
}
```

- [ ] **Step 2: Create `internal/provider/event_resource.go`**

Remove stub `ResendEventResource`, then create:

```go
package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ resource.Resource = &EventResource{}
var _ resource.ResourceWithImportState = &EventResource{}
var _ resource.ResourceWithConfigure = &EventResource{}

type EventResource struct{ client *resend.Client }

type EventResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Schema types.String `tfsdk:"schema"`
}

func ResendEventResource() resource.Resource { return &EventResource{} }

func (r *EventResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event"
}

func (r *EventResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend event. Cannot use the resend: name prefix.",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, Description: "Event UUID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":   schema.StringAttribute{Required: true, Description: "Event name. Cannot start with resend:."},
			"schema": schema.StringAttribute{Optional: true, Computed: true, Description: "JSON object defining event payload key/type pairs.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *EventResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*resend.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func parseSchema(s types.String) any {
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(s.ValueString()), &v); err != nil {
		return nil
	}
	return v
}

func (r *EventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateEvent(ctx, resend.CreateEventRequest{
		Name:   plan.Name.ValueString(),
		Schema: parseSchema(plan.Schema),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Event", err.Error())
		return
	}
	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	e, err := r.client.GetEvent(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Event", err.Error())
		return
	}
	state.Name = types.StringValue(e.Name)
	if e.Schema != "" {
		state.Schema = types.StringValue(e.Schema)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state EventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateEvent(ctx, state.ID.ValueString(), resend.UpdateEventRequest{
		Name:   plan.Name.ValueString(),
		Schema: parseSchema(plan.Schema),
	}); err != nil {
		resp.Diagnostics.AddError("Error Updating Event", err.Error())
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEvent(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Event", err.Error())
	}
}

func (r *EventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

- [ ] **Step 3: Create `internal/provider/event_resource_test.go`**

```go
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
```

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no errors. The stub file `template_resource.go` now has no stubs left — delete it and confirm no orphan stubs remain.

- [ ] **Step 5: Commit**

```bash
git add resend/event.go internal/provider/event_resource.go internal/provider/event_resource_test.go
git commit -m "feat: add resend_event resource"
```

---

## Task 10: CI/CD and release config

**Files:**
- Create: `.github/workflows/test.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yml`

- [ ] **Step 1: Create `.github/workflows/test.yml`**

```yaml
name: Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go test ./... -v
```

- [ ] **Step 2: Create `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        id: import_gpg
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

- [ ] **Step 3: Create `.goreleaser.yml`**

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
    flags:
      - -trimpath
    ldflags:
      - "-s -w -X main.version={{.Version}}"
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    binary: "{{ .ProjectName }}_v{{ .Version }}"

archives:
  - format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "{{ .ProjectName }}_{{ .Version }}_SHA256SUMS"
  algorithm: sha256

signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"

release:
  draft: true

changelog:
  skip: true
```

- [ ] **Step 4: Commit**

```bash
git add .github/ .goreleaser.yml
git commit -m "ci: add test workflow and goreleaser release config"
```

---

## Task 11: CLAUDE.md, README.md, and examples

**Files:**
- Create: `CLAUDE.md`
- Create: `README.md`
- Create: `examples/provider/main.tf`
- Create: `examples/resources/resend_template/resource.tf`

- [ ] **Step 1: Create `CLAUDE.md`**

```markdown
# CLAUDE.md

Instructions for Claude Code working in this repository.

## Overview

Public Terraform provider for the Resend email API. Published on the Terraform Registry under `realianx/resend`.

## Tech Stack

- Go 1.25+
- `hashicorp/terraform-plugin-framework` (TPF)
- `goreleaser` for registry publishing
- `tfplugindocs` for documentation generation

## Project Structure

```
terraform-provider-resend/
├── resend/                    # HTTP client (one file per API domain)
│   ├── client.go              # base client, Bearer auth, HTTPError
│   ├── template.go
│   ├── segment.go
│   ├── topic.go
│   ├── automation.go
│   └── event.go
├── internal/provider/
│   ├── provider.go            # config (api_key) + resource registration
│   ├── template_resource.go
│   ├── segment_resource.go
│   ├── topic_resource.go
│   ├── automation_resource.go
│   └── event_resource.go
├── .github/workflows/
├── examples/
├── docs/                      # generated by tfplugindocs
├── main.go
├── go.mod
├── .goreleaser.yml
└── GNUmakefile
```

## Key Commands

```bash
go build ./...           # build
make test                # unit tests
make testacc             # acceptance tests (requires RESEND_API_KEY)
make docs                # generate docs
goreleaser release --snapshot --clean  # dry-run release
```

## Conventions

- One resource = one file in `internal/provider/`
- One API domain = one file in `resend/`
- Acceptance tests in `internal/provider/*_resource_test.go`
- `RESEND_API_KEY` environment variable for provider config
- `isNotFound()` helper in `template_resource.go` for 404 detection

## API

- Base URL: `https://api.resend.com`
- Auth: `Authorization: Bearer <api_key>`
- Docs: https://resend.com/docs/api-reference/introduction

## Publishing

1. Add GPG key as GitHub secret `GPG_PRIVATE_KEY`
2. Push tag `v0.1.0` → goreleaser publishes automatically
3. Usage: `source = "realianx/resend"` in `required_providers`
```

- [ ] **Step 2: Create `README.md`**

```markdown
# terraform-provider-resend

Terraform provider for [Resend](https://resend.com) — manage email templates, segments, topics, automations, and events as infrastructure.

## Requirements

- Terraform >= 1.0
- Go >= 1.25 (for development)
- Resend API key

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
  api_key = var.resend_api_key  # or set RESEND_API_KEY
}

resource "resend_template" "welcome" {
  name    = "Welcome Email"
  html    = "<p>Welcome, {{name}}!</p>"
  subject = "Welcome!"
}

resource "resend_segment" "active_users" {
  name = "Active Users"
}

resource "resend_topic" "newsletter" {
  name                 = "Newsletter"
  default_subscription = "opt_in"
}

resource "resend_automation" "onboarding" {
  name   = "Onboarding Flow"
  status = "enabled"
}

resource "resend_event" "user_signup" {
  name   = "user.signup"
  schema = jsonencode({ email = "string", plan = "string" })
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `resend_template` | Email template (CRUD + publish) |
| `resend_segment` | Contact segment (create/delete, no update) |
| `resend_topic` | Subscription topic (CRUD) |
| `resend_automation` | Automation workflow (CRUD, steps as JSON) |
| `resend_event` | Custom event (CRUD) |

## Development

```bash
go build ./...
make test
RESEND_API_KEY=re_xxx TF_ACC=1 make testacc
```

## License

MIT
```

- [ ] **Step 3: Create `examples/provider/main.tf`**

```hcl
provider "resend" {
  # api_key = "re_xxx"  # or set RESEND_API_KEY env var
}
```

- [ ] **Step 4: Create `examples/resources/resend_template/resource.tf`**

```hcl
resource "resend_template" "example" {
  name      = "Welcome Email"
  html      = "<p>Hello {{name}}, welcome!</p>"
  subject   = "Welcome!"
  from      = "hello@example.com"
  published = true
}
```

- [ ] **Step 5: Commit**

```bash
git add CLAUDE.md README.md examples/
git commit -m "docs: add CLAUDE.md, README.md, and examples"
```

---

## Task 12: Final verification

- [ ] **Step 1: Full build**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 2: Unit tests**

```bash
go test ./resend/... -v
```

Expected: all `PASS`

- [ ] **Step 3: Goreleaser dry-run**

```bash
goreleaser release --snapshot --clean
```

Expected: binaries built in `dist/` for linux/darwin/windows × amd64/arm64.

- [ ] **Step 4: (Optional) Acceptance tests**

```bash
RESEND_API_KEY=re_xxx TF_ACC=1 go test ./internal/provider/... -v -timeout 10m
```

Expected: all resources created, updated, and destroyed without errors.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final cleanup and verification"
```

---

## Publishing Checklist

- [ ] Create GitHub repo `terraform-provider-resend` under `realianx`
- [ ] Add `GPG_PRIVATE_KEY` secret to repo settings
- [ ] Register namespace `realianx` on registry.terraform.io
- [ ] Push code + tag: `git tag v0.1.0 && git push origin v0.1.0 --tags`
- [ ] Approve draft release on GitHub after goreleaser runs
