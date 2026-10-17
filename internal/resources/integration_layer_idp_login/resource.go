package integration_layer_idp_login

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/integrationlayer"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

const (
	MatchExternalID = "externalId"
	MatchEmail      = "email"
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client     *integrationlayer.ByProjectKeyRequestBuilder
	projectKey string
	mutex      *utils.MutexKV
}

type IdpLogin struct {
	ID                    types.String `tfsdk:"id"`
	Issuer                types.String `tfsdk:"issuer"`
	TokenEndpoint         types.String `tfsdk:"token_endpoint"`
	RevocationEndpoint    types.String `tfsdk:"revocation_endpoint"`
	JwksURI               types.String `tfsdk:"jwks_uri"`
	AuthorizationEndpoint types.String `tfsdk:"authorization_endpoint"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecret          types.String `tfsdk:"client_secret"`
	HasClientSecret       types.Bool   `tfsdk:"has_client_secret"`
	RedirectURI           types.String `tfsdk:"redirect_uri"`
	MatchStrategy         types.String `tfsdk:"match_strategy"`
	ExternalIDClaim       types.String `tfsdk:"external_id_claim"`
	EmailClaim            types.String `tfsdk:"email_claim"`
	FirstNameClaim        types.String `tfsdk:"first_name_claim"`
	LastNameClaim         types.String `tfsdk:"last_name_claim"`
	PhoneClaim            types.String `tfsdk:"phone_claim"`
}

// The token claim each Customer field is read from by default: the OIDC
// standard claims, which are also what the commercetools CLI proposes.
var defaultClaims = map[string]struct{ field, claim string }{
	"external_id_claim": {"external ID", "sub"},
	"email_claim":       {"email address", "email"},
	"first_name_claim":  {"first name", "given_name"},
	"last_name_claim":   {"last name", "family_name"},
	"phone_claim":       {"phone number", "phone_number"},
}

// An absolute http(s) URL without credentials in it, which is what the API
// accepts for the endpoints it calls.
var endpointURL = stringvalidator.RegexMatches(
	regexp.MustCompile(`^https?://[^/@\s]+(/\S*)?$`),
	"must be an absolute http(s) URL without credentials",
)

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_layer_idp_login"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	endpoint := func(description string) schema.StringAttribute {
		return schema.StringAttribute{
			Description: description,
			Required:    true,
			Validators:  []validator.String{endpointURL},
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Customer login on the Commerce Integration Layer through an external OpenID Connect " +
			"identity provider. A customer who signs in there gets a Commerce Integration Layer session, " +
			"matched to a commercetools Customer by `match_strategy`.\n\n" +
			"A Project has at most one identity provider, so define this resource once per Project. Destroying " +
			"it disables the login.\n\n" +
			"The API never returns the client secret. Terraform keeps the configured value in state and cannot " +
			"detect that it was changed outside of Terraform. After an import, the secret stays as stored until " +
			"`client_secret` is set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The key of the Project the login belongs to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				Description: "Issuer of the identity provider's tokens, as in their `iss` claim.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"token_endpoint":      endpoint("The identity provider's token endpoint."),
			"revocation_endpoint": endpoint("The identity provider's token revocation endpoint."),
			"jwks_uri":            endpoint("Where the identity provider publishes the keys its tokens are signed with."),
			"authorization_endpoint": schema.StringAttribute{
				Description: "The identity provider's authorization endpoint, for a login started by the " +
					"Commerce Integration Layer.",
				Optional:   true,
				Validators: []validator.String{endpointURL},
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID of the application registered with the identity provider.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client secret of that application. Required when the login is first " +
					"created. Leave it out afterwards to keep the stored secret.",
				Optional:   true,
				Sensitive:  true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"has_client_secret": schema.BoolAttribute{
				Description: "Whether the Commerce Integration Layer holds a client secret.",
				Computed:    true,
			},
			"redirect_uri": schema.StringAttribute{
				Description: "Redirect URI registered with the identity provider.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"match_strategy": schema.StringAttribute{
				MarkdownDescription: "How a signed-in customer is matched to a commercetools Customer: " +
					"`externalId` by the `external_id` claim, `email` by the email address.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(MatchExternalID, MatchEmail),
				},
			},
		},
	}
	for name, d := range defaultClaims {
		resp.Schema.Attributes[name] = schema.StringAttribute{
			MarkdownDescription: "Token claim the Customer's " + d.field + " is read from. Defaults to `" +
				d.claim + "`.",
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString(d.claim),
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		}
	}
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
	var plan IdpLogin
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.write(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating integration layer IdP login", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IdpLogin
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Main().Idp().Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading integration layer IdP login", err.Error())
		return
	}
	current := NewIdpLoginFromNative(r.projectKey, res)
	current.ClientSecret = state.ClientSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IdpLogin
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.write(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating integration layer IdP login", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}

func (r *Resource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)

	err := r.client.Main().Idp().Delete().Execute(ctx)
	if err != nil && !utils.IsResourceNotFoundError(err) {
		resp.Diagnostics.AddError("Error deleting integration layer IdP login", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) write(ctx context.Context, plan IdpLogin) (IdpLogin, error) {
	r.mutex.Lock(utils.IntegrationLayerLock)
	defer r.mutex.Unlock(utils.IntegrationLayerLock)

	res, err := r.client.Main().Idp().Put(plan.draft()).Execute(ctx)
	if err != nil {
		return IdpLogin{}, err
	}
	current := NewIdpLoginFromNative(r.projectKey, res)
	current.ClientSecret = plan.ClientSecret
	return current, nil
}

// draft sends an empty secret when none is configured, which keeps the
// stored one; the API rejects it only when it holds none yet.
func (l IdpLogin) draft() integrationlayer.IdpConfigInput {
	return integrationlayer.IdpConfigInput{
		Issuer:                l.Issuer.ValueString(),
		TokenEndpoint:         l.TokenEndpoint.ValueString(),
		RevocationEndpoint:    l.RevocationEndpoint.ValueString(),
		JwksUri:               l.JwksURI.ValueString(),
		AuthorizationEndpoint: utils.OptionalString(l.AuthorizationEndpoint),
		ClientId:              l.ClientID.ValueString(),
		RedirectUri:           l.RedirectURI.ValueString(),
		MatchStrategy:         integrationlayer.IdpMatchStrategy(l.MatchStrategy.ValueString()),
		ClientSecret:          l.ClientSecret.ValueString(),
		Claims: integrationlayer.IdpClaimMapping{
			ExternalId: l.ExternalIDClaim.ValueString(),
			Email:      l.EmailClaim.ValueString(),
			FirstName:  l.FirstNameClaim.ValueString(),
			LastName:   l.LastNameClaim.ValueString(),
			Phone:      l.PhoneClaim.ValueString(),
		},
	}
}

// NewIdpLoginFromNative maps the redacted configuration the API answers with.
// The client secret is not in it, so the caller carries it over.
func NewIdpLoginFromNative(projectKey string, c *integrationlayer.IdpConfig) IdpLogin {
	return IdpLogin{
		ID:                    types.StringValue(projectKey),
		Issuer:                types.StringValue(c.Issuer),
		TokenEndpoint:         types.StringValue(c.TokenEndpoint),
		RevocationEndpoint:    types.StringValue(c.RevocationEndpoint),
		JwksURI:               types.StringValue(c.JwksUri),
		AuthorizationEndpoint: utils.FromOptionalString(c.AuthorizationEndpoint),
		ClientID:              types.StringValue(c.ClientId),
		ClientSecret:          types.StringNull(),
		HasClientSecret:       types.BoolValue(c.HasClientSecret),
		RedirectURI:           types.StringValue(c.RedirectUri),
		MatchStrategy:         types.StringValue(string(c.MatchStrategy)),
		ExternalIDClaim:       types.StringValue(c.Claims.ExternalId),
		EmailClaim:            types.StringValue(c.Claims.Email),
		FirstNameClaim:        types.StringValue(c.Claims.FirstName),
		LastNameClaim:         types.StringValue(c.Claims.LastName),
		PhoneClaim:            types.StringValue(c.Claims.Phone),
	}
}
