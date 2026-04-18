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
	CreatedAt           types.String `tfsdk:"created_at"`
}

func ResendTopicResource() resource.Resource { return &TopicResource{} }

func (r *TopicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topic"
}

func (r *TopicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Resend topic.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Topic UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Topic name. Max 50 characters.",
			},
			"default_subscription": schema.StringAttribute{
				Required:      true,
				Description:   "opt_in or opt_out. Immutable after creation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Topic description. Max 200 characters.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"visibility": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "public or private.",
				Default:     stringdefault.StaticString("private"),
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 creation timestamp.",
			},
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

	t, err := r.client.GetTopic(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Topic After Create", err.Error())
		return
	}
	plan.CreatedAt = types.StringValue(t.CreatedAt)

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
	state.CreatedAt = types.StringValue(t.CreatedAt)
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
	plan.CreatedAt = state.CreatedAt
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
