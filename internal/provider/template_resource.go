package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	Variables types.List   `tfsdk:"variables"`
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
			"id":       schema.StringAttribute{Computed: true, Description: "Template UUID.", PlanModifiers: computed()},
			"name":     schema.StringAttribute{Required: true, Description: "Template name."},
			"html":     schema.StringAttribute{Required: true, Description: "HTML body."},
			"alias":    schema.StringAttribute{Optional: true, Computed: true, Description: "Template alias.", PlanModifiers: computed()},
			"from":     schema.StringAttribute{Optional: true, Computed: true, Description: "Default sender.", PlanModifiers: computed()},
			"subject":  schema.StringAttribute{Optional: true, Computed: true, Description: "Default subject.", PlanModifiers: computed()},
			"reply_to": schema.StringAttribute{Optional: true, Computed: true, Description: "Default reply-to.", PlanModifiers: computed()},
			"text":     schema.StringAttribute{Optional: true, Computed: true, Description: "Plain-text body.", PlanModifiers: computed()},
			"published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Set true to publish. Resend has no unpublish endpoint; this only triggers a publish call.",
			},
			"variables": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "List of variable names used in the template (required when using {{{var}}} placeholders).",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
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
		Name:      plan.Name.ValueString(),
		HTML:      plan.HTML.ValueString(),
		Alias:     plan.Alias.ValueString(),
		From:      plan.From.ValueString(),
		Subject:   plan.Subject.ValueString(),
		ReplyTo:   plan.ReplyTo.ValueString(),
		Text:      plan.Text.ValueString(),
		Variables: listToStrings(ctx, plan.Variables),
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
		resp.Diagnostics.AddError("Error Reading Template After Create", err.Error())
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
		Name:      plan.Name.ValueString(),
		HTML:      plan.HTML.ValueString(),
		Alias:     plan.Alias.ValueString(),
		From:      plan.From.ValueString(),
		Subject:   plan.Subject.ValueString(),
		ReplyTo:   plan.ReplyTo.ValueString(),
		Text:      plan.Text.ValueString(),
		Variables: listToStrings(ctx, plan.Variables),
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
	elems := make([]attr.Value, len(t.Variables))
	for i, v := range t.Variables {
		elems[i] = types.StringValue(v)
	}
	m.Variables, _ = types.ListValue(types.StringType, elems)
}

func listToStrings(ctx context.Context, l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var out []string
	_ = l.ElementsAs(ctx, &out, false)
	return out
}

func isNotFound(err error) bool {
	var e *resend.HTTPError
	return errors.As(err, &e) && e.StatusCode == 404
}
