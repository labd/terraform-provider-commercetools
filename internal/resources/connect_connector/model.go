package connect_connector

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/connect"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type Connector struct {
	ID               types.String `tfsdk:"id"`
	Key              types.String `tfsdk:"key"`
	Version          types.Int64  `tfsdk:"version"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Creator          *Creator     `tfsdk:"creator"`
	RepositoryURL    types.String `tfsdk:"repository_url"`
	RepositoryTag    types.String `tfsdk:"repository_tag"`
	IntegrationTypes types.Set    `tfsdk:"integration_types"`
	SupportedRegions types.Set    `tfsdk:"supported_regions"`
	PrivateProjects  types.Set    `tfsdk:"private_projects"`
	DocumentationURL types.String `tfsdk:"documentation_url"`
	Publish          types.Bool   `tfsdk:"publish"`
	PublishTimeout   types.String `tfsdk:"publish_timeout"`
	Status           types.String `tfsdk:"status"`
	HasChanges       types.Bool   `tfsdk:"has_changes"`
	PublishedVersion types.Int64  `tfsdk:"published_version"`
}

type Creator struct {
	Email            types.String `tfsdk:"email"`
	Name             types.String `tfsdk:"name"`
	Title            types.String `tfsdk:"title"`
	Company          types.String `tfsdk:"company"`
	LogoURL          types.String `tfsdk:"logo_url"`
	SupportURL       types.String `tfsdk:"support_url"`
	NoOfContributors types.Int64  `tfsdk:"no_of_contributors"`
}

// creator is the configured creator, or an empty one after an import.
func (c Connector) creator() Creator {
	if c.Creator == nil {
		return Creator{}
	}
	return *c.Creator
}

func (c Connector) draft(ctx context.Context) (connect.ConnectorStagedDraft, diag.Diagnostics) {
	var diags diag.Diagnostics
	creator := c.creator()
	draft := connect.ConnectorStagedDraft{
		Key:         utils.OptionalString(c.Key),
		Name:        c.Name.ValueString(),
		Description: utils.OptionalString(c.Description),
		Creator: connect.Creator{
			Email:            creator.Email.ValueString(),
			Name:             utils.OptionalString(creator.Name),
			Title:            utils.OptionalString(creator.Title),
			Company:          utils.OptionalString(creator.Company),
			LogoUrl:          utils.OptionalString(creator.LogoURL),
			SupportUrl:       utils.OptionalString(creator.SupportURL),
			NoOfContributors: utils.OptionalInt(creator.NoOfContributors),
		},
		Repository: connect.Repository{
			Url: c.RepositoryURL.ValueString(),
			Tag: c.RepositoryTag.ValueString(),
		},
		DocumentationUrl: utils.OptionalString(c.DocumentationURL),
	}
	draft.IntegrationTypes = toStrings[connect.IntegrationType](ctx, c.IntegrationTypes, &diags)
	draft.SupportedRegions = toStrings[connect.Region](ctx, c.SupportedRegions, &diags)
	draft.PrivateProjects = toStrings[string](ctx, c.PrivateProjects, &diags)
	return draft, diags
}

// updateActions are the actions that change the staged connector from state
// into c.
func (c Connector) updateActions(ctx context.Context, state Connector) ([]connect.ConnectorUpdateAction, diag.Diagnostics) {
	var diags diag.Diagnostics
	var actions []connect.ConnectorUpdateAction
	planCreator, stateCreator := c.creator(), state.creator()
	setString := func(planned, current types.String, action func(string) connect.ConnectorUpdateAction) {
		if !planned.Equal(current) {
			actions = append(actions, action(planned.ValueString()))
		}
	}

	setString(c.Name, state.Name, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetNameAction{Name: v}
	})
	setString(c.Description, state.Description, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetDescriptionAction{Description: v}
	})
	setString(c.DocumentationURL, state.DocumentationURL, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetDocumentationUrlAction{DocumentationUrl: v}
	})
	setString(planCreator.Email, stateCreator.Email, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorEmailAction{CreatorEmail: v}
	})
	setString(planCreator.Name, stateCreator.Name, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorNameAction{CreatorName: v}
	})
	setString(planCreator.Title, stateCreator.Title, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorTitleAction{CreatorTitle: v}
	})
	setString(planCreator.Company, stateCreator.Company, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorCompanyAction{CreatorCompany: v}
	})
	setString(planCreator.LogoURL, stateCreator.LogoURL, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorLogoAction{LogoUrl: v}
	})
	setString(planCreator.SupportURL, stateCreator.SupportURL, func(v string) connect.ConnectorUpdateAction {
		return connect.ConnectorSetCreatorSupportUrlAction{CreatorSupportUrl: v}
	})
	if !planCreator.NoOfContributors.Equal(stateCreator.NoOfContributors) {
		actions = append(actions, connect.ConnectorSetCreatorNoOfContributorsAction{
			CreatorNoOfContributors: int(planCreator.NoOfContributors.ValueInt64()),
		})
	}
	if !c.RepositoryURL.Equal(state.RepositoryURL) || !c.RepositoryTag.Equal(state.RepositoryTag) {
		actions = append(actions, connect.ConnectorSetRepositoryAction{
			Url: c.RepositoryURL.ValueString(),
			Tag: c.RepositoryTag.ValueString(),
		})
	}

	planned := toStrings[connect.IntegrationType](ctx, c.IntegrationTypes, &diags)
	if !sameElements(planned, toStrings[connect.IntegrationType](ctx, state.IntegrationTypes, &diags)) {
		actions = append(actions, connect.ConnectorSetIntegrationTypesAction{IntegrationTypes: planned})
	}
	regions := toStrings[connect.Region](ctx, c.SupportedRegions, &diags)
	if !sameElements(regions, toStrings[connect.Region](ctx, state.SupportedRegions, &diags)) {
		actions = append(actions, connect.ConnectorSetSupportedRegionsAction{Regions: regions})
	}

	projects := toStrings[string](ctx, c.PrivateProjects, &diags)
	current := toStrings[string](ctx, state.PrivateProjects, &diags)
	for _, project := range current {
		if !slices.Contains(projects, project) {
			actions = append(actions, connect.ConnectorRemovePrivateProjectAction{PrivateProject: project})
		}
	}
	for _, project := range projects {
		if !slices.Contains(current, project) {
			actions = append(actions, connect.ConnectorAddPrivateProjectAction{PrivateProject: project})
		}
	}
	return actions, diags
}

// NewConnectorFromNative fills the model from the staged connector, in the
// shape of prior, the plan or the state. publishedVersion is the version of
// the published connector, if there is one. A value the configuration leaves
// out that the API returns empty stays left out.
func NewConnectorFromNative(ctx context.Context, res *connect.ConnectorStaged, publishedVersion *int, prior Connector) (Connector, diag.Diagnostics) {
	var diags diag.Diagnostics
	current := prior
	current.ID = types.StringValue(res.ID)
	current.Key = utils.FromOptionalString(res.Key)
	current.Version = types.Int64Value(int64(res.Version))
	current.Name = types.StringValue(res.Name)
	current.Description = optionalText(&res.Description, prior.Description)
	current.DocumentationURL = optionalText(res.DocumentationUrl, prior.DocumentationURL)
	current.RepositoryURL = types.StringValue(res.Repository.Url)
	current.RepositoryTag = types.StringValue(res.Repository.Tag)
	current.Status = types.StringValue(res.Status)
	current.HasChanges = types.BoolValue(res.HasChanges)
	current.PublishedVersion = utils.FromOptionalInt(publishedVersion)

	priorCreator := prior.creator()
	current.Creator = &Creator{
		Email:            types.StringValue(res.Creator.Email),
		Name:             optionalText(res.Creator.Name, priorCreator.Name),
		Title:            optionalText(res.Creator.Title, priorCreator.Title),
		Company:          optionalText(res.Creator.Company, priorCreator.Company),
		LogoURL:          optionalText(res.Creator.LogoUrl, priorCreator.LogoURL),
		SupportURL:       optionalText(res.Creator.SupportUrl, priorCreator.SupportURL),
		NoOfContributors: utils.FromOptionalInt(res.Creator.NoOfContributors),
	}

	current.IntegrationTypes = fromStrings(ctx, res.IntegrationTypes, prior.IntegrationTypes, &diags)
	current.SupportedRegions = fromStrings(ctx, res.SupportedRegions, prior.SupportedRegions, &diags)
	current.PrivateProjects = fromStrings(ctx, res.PrivateProjects, prior.PrivateProjects, &diags)
	if current.Publish.IsNull() || current.Publish.IsUnknown() {
		current.Publish = types.BoolValue(false)
	}
	return current, diags
}

// optionalText is the value, or null when it is empty and prior is null.
func optionalText(value *string, prior types.String) types.String {
	if value == nil || (*value == "" && prior.IsNull()) {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func toStrings[T ~string](ctx context.Context, set types.Set, diags *diag.Diagnostics) []T {
	out := []T{}
	if set.IsNull() || set.IsUnknown() {
		return out
	}
	var values []string
	diags.Append(set.ElementsAs(ctx, &values, false)...)
	for _, v := range values {
		out = append(out, T(v))
	}
	return out
}

func fromStrings[T ~string](ctx context.Context, values []T, prior types.Set, diags *diag.Diagnostics) types.Set {
	if len(values) == 0 && prior.IsNull() {
		return types.SetNull(types.StringType)
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}
	set, d := types.SetValueFrom(ctx, types.StringType, out)
	diags.Append(d...)
	return set
}

func sameElements[T ~string](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}
