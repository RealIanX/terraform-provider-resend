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
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Event UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Event name. Cannot start with resend:.",
			},
			"schema": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "JSON object defining event payload key/type pairs.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
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

func parseEventSchema(s types.String) any {
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
		Schema: parseEventSchema(plan.Schema),
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
	if len(e.Schema) > 0 && string(e.Schema) != "null" {
		state.Schema = types.StringValue(string(e.Schema))
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
		Schema: parseEventSchema(plan.Schema),
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
