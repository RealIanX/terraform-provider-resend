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
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func ResendAutomationResource() resource.Resource { return &AutomationResource{} }

func (r *AutomationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

func (r *AutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend automation. steps and connections are JSON strings (use jsonencode()).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Automation UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Automation name.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "enabled or disabled.",
				Default:     stringdefault.StaticString("disabled"),
			},
			"steps": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "JSON array of automation steps. Must be provided with connections.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connections": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "JSON array of step connections. Must be provided with steps.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 last-update timestamp.",
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

	a, err := r.client.GetAutomation(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Automation After Create", err.Error())
		return
	}
	plan.CreatedAt = types.StringValue(a.CreatedAt)
	plan.UpdatedAt = types.StringValue(a.UpdatedAt)

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
	if len(a.Steps) > 0 && string(a.Steps) != "null" {
		state.Steps = types.StringValue(string(a.Steps))
	}
	if len(a.Connections) > 0 && string(a.Connections) != "null" {
		state.Connections = types.StringValue(string(a.Connections))
	}
	state.CreatedAt = types.StringValue(a.CreatedAt)
	state.UpdatedAt = types.StringValue(a.UpdatedAt)
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
	plan.CreatedAt = state.CreatedAt

	a, err := r.client.GetAutomation(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Automation After Update", err.Error())
		return
	}
	plan.UpdatedAt = types.StringValue(a.UpdatedAt)

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
