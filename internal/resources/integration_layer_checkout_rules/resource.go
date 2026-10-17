package integration_layer_checkout_rules

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/integrationlayer"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                   = &Resource{}
	_ resource.ResourceWithConfigure      = &Resource{}
	_ resource.ResourceWithImportState    = &Resource{}
	_ resource.ResourceWithValidateConfig = &Resource{}
)

const (
	Complete    = "COMPLETE"
	PaymentOnly = "PAYMENT_ONLY"
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client     *integrationlayer.ByProjectKeyRequestBuilder
	projectKey string
	mutex      *utils.MutexKV
}

type CheckoutRules struct {
	ID    types.String `tfsdk:"id"`
	Rules []Rule       `tfsdk:"rule"`
}

type Rule struct {
	Countries      []types.String `tfsdk:"countries"`
	Stores         []types.String `tfsdk:"stores"`
	ApplicationKey types.String   `tfsdk:"application_key"`
	Mode           types.String   `tfsdk:"mode"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_checkout_rules"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The rules that pick the commercetools Checkout application a cart on the Commerce " +
			"Integration Layer checks out with. The first rule matching the cart's country and store wins.\n\n" +
			"A Project has one list of rules, so define this resource once per Project. The API cannot remove " +
			"the rules: destroying the resource only removes it from the state, and the rules stay in force.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the rules belong to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				MarkdownDescription: "A rule, in order of precedence. A rule without `countries` and `stores` " +
					"matches every cart, so it can only be the last.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"countries": schema.ListAttribute{
							MarkdownDescription: "Countries, as ISO 3166-1 alpha-2 codes, of which the cart's " +
								"must be one. Leave out to match any country.",
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]{2}$`),
										"must be an uppercase ISO 3166-1 alpha-2 code"),
								),
							},
						},
						"stores": schema.ListAttribute{
							MarkdownDescription: "Keys of the Stores of which the cart's must be one. Leave out " +
								"to match any Store.",
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9_-]+$`),
										"must be a Store key"),
								),
							},
						},
						"application_key": schema.StringAttribute{
							Description: "Key of the commercetools Checkout application to check out with.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: "`COMPLETE` for the full checkout, `PAYMENT_ONLY` for payment " +
								"only.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(Complete, PaymentOnly),
							},
						},
					},
				},
			},
		},
	}
}

// ValidateConfig rejects a catch-all rule before the last position, where it
// would shadow every rule after it.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config CheckoutRules
	// A rule list that is not known yet, as with a dynamic block over a
	// variable, is checked by the API at apply time instead.
	if diags := req.Config.Get(ctx, &config); diags.HasError() {
		return
	}
	resp.Diagnostics.Append(validateOrder(config.Rules)...)
}

func validateOrder(rules []Rule) diag.Diagnostics {
	var diags diag.Diagnostics
	for i, rule := range rules[:max(len(rules)-1, 0)] {
		if rule.Countries == nil && rule.Stores == nil {
			diags.AddAttributeError(
				path.Root("rule").AtListIndex(i),
				"Catch-all checkout rule before the last",
				fmt.Sprintf("Rule %d has neither countries nor stores, so it matches every cart and no rule "+
					"after it would ever apply. Move it to the end.", i+1),
			)
		}
	}
	return diags
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*utils.ProviderData)
	if !ok {
		return
	}
	r.client = data.IntegrationLayerClient
	r.projectKey = data.ProjectKey
	r.mutex = data.Mutex
	if r.client == nil {
		resp.Diagnostics.AddError("Commerce Integration Layer URL unknown",
			"Set integration_layer_url on the provider, or CTP_INTEGRATION_LAYER_URL, when api_url "+
				"is not a commercetools.com host the URL can be derived from.")
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CheckoutRules
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error creating integration layer checkout rules", err.Error())
		return
	}
	plan.ID = types.StringValue(r.projectKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CheckoutRules
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.Main().Settings().Get().Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration layer checkout rules", err.Error())
		return
	}
	// A Project whose checkout was never configured has no rules to manage.
	if view.Project.Checkout == nil || len(view.Project.Checkout.Rules) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	current := NewCheckoutRulesFromNative(r.projectKey, *view.Project.Checkout)
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CheckoutRules
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.write(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error updating integration layer checkout rules", err.Error())
		return
	}
	plan.ID = types.StringValue(r.projectKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete only removes the rules from the state. The API takes no empty list,
// and a Project without rules cannot check out, so they stay in force.
func (r *Resource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Integration layer checkout rules are kept",
		"The Commerce Integration Layer cannot remove a Project's checkout rules. They are removed from the "+
			"Terraform state only, and stay in force on the Project.",
	)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) write(ctx context.Context, plan CheckoutRules) error {
	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)
	_, err := r.client.Main().Checkout().Put(plan.draft()).Execute(ctx)
	return err
}

func (c CheckoutRules) draft() integrationlayer.CheckoutConfig {
	rules := make([]integrationlayer.CheckoutRule, 0, len(c.Rules))
	for _, rule := range c.Rules {
		rules = append(rules, integrationlayer.CheckoutRule{
			Countries:      stringsOf(rule.Countries),
			Stores:         stringsOf(rule.Stores),
			ApplicationKey: rule.ApplicationKey.ValueString(),
			Mode:           integrationlayer.CheckoutMode(rule.Mode.ValueString()),
		})
	}
	return integrationlayer.CheckoutConfig{Rules: rules}
}

func NewCheckoutRulesFromNative(projectKey string, config integrationlayer.CheckoutConfig) CheckoutRules {
	rules := make([]Rule, 0, len(config.Rules))
	for _, rule := range config.Rules {
		rules = append(rules, Rule{
			Countries:      values(rule.Countries),
			Stores:         values(rule.Stores),
			ApplicationKey: types.StringValue(rule.ApplicationKey),
			Mode:           types.StringValue(string(rule.Mode)),
		})
	}
	return CheckoutRules{ID: types.StringValue(projectKey), Rules: rules}
}

func stringsOf(in []types.String) []string {
	if in == nil {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.ValueString())
	}
	return out
}

// values keeps an absent list absent, so a rule that matches any country
// reads back the way it is configured.
func values(in []string) []types.String {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.String, 0, len(in))
	for _, v := range in {
		out = append(out, types.StringValue(v))
	}
	return out
}
