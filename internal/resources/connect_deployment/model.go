package connect_deployment

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/connect"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type Deployment struct {
	ID                          types.String  `tfsdk:"id"`
	Key                         types.String  `tfsdk:"key"`
	Version                     types.Int64   `tfsdk:"version"`
	ConnectorKey                types.String  `tfsdk:"connector_key"`
	ConnectorID                 types.String  `tfsdk:"connector_id"`
	ConnectorVersion            types.Int64   `tfsdk:"connector_version"`
	ConnectorStaged             types.Bool    `tfsdk:"connector_staged"`
	Type                        types.String  `tfsdk:"type"`
	Region                      types.String  `tfsdk:"region"`
	Applications                []Application `tfsdk:"application"`
	GlobalStandardConfiguration types.Map     `tfsdk:"global_standard_configuration"`
	GlobalSecuredConfiguration  types.Map     `tfsdk:"global_secured_configuration"`
	SkipScriptsOnRedeploy       types.Bool    `tfsdk:"skip_scripts_on_redeploy"`
	DeployTimeout               types.String  `tfsdk:"deploy_timeout"`
	Status                      types.String  `tfsdk:"status"`
}

type Application struct {
	Name                  types.String `tfsdk:"name"`
	StandardConfiguration types.Map    `tfsdk:"standard_configuration"`
	SecuredConfiguration  types.Map    `tfsdk:"secured_configuration"`
	Schedule              types.String `tfsdk:"schedule"`
	URL                   types.String `tfsdk:"url"`
	Topic                 types.String `tfsdk:"topic"`
}

func (d Deployment) draft(ctx context.Context) (connect.DeploymentDraft, diag.Diagnostics) {
	configurations, diags := d.configurationValues(ctx)
	global, globalDiags := d.globalConfiguration(ctx)
	diags.Append(globalDiags...)

	draft := connect.DeploymentDraft{
		Key:                 utils.OptionalString(d.Key),
		Type:                utils.OptionalString(d.Type),
		Region:              d.Region.ValueString(),
		Configurations:      configurations,
		GlobalConfiguration: global,
		Connector: connect.ConnectorReference{
			ID:     utils.OptionalString(d.ConnectorID),
			Key:    utils.OptionalString(d.ConnectorKey),
			Staged: optionalBool(d.ConnectorStaged),
		},
	}
	if !d.ConnectorVersion.IsNull() && !d.ConnectorVersion.IsUnknown() {
		draft.Connector.Version = utils.Ref(int(d.ConnectorVersion.ValueInt64()))
	}
	return draft, diags
}

// redeploy is the action that moves the deployment from state to d. A raised
// connector version moves it to the connector's latest version, the only
// one a redeploy can move it to.
func (d Deployment) redeploy(ctx context.Context, state Deployment) (connect.DeploymentUpdateAction, diag.Diagnostics) {
	configurations, diags := d.configurationValues(ctx)
	global, globalDiags := d.globalConfiguration(ctx)
	diags.Append(globalDiags...)
	action := connect.DeploymentRedeployAction{
		SkipScripts:         optionalBool(d.SkipScriptsOnRedeploy),
		ConfigurationValues: configurations,
		GlobalConfiguration: global,
	}
	if !d.ConnectorVersion.IsUnknown() && !d.ConnectorVersion.Equal(state.ConnectorVersion) {
		action.UpdateConnector = utils.Ref(true)
	}
	return action, diags
}

func (d Deployment) configurationValues(ctx context.Context) ([]connect.DeploymentConfigurationApplication, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make([]connect.DeploymentConfigurationApplication, 0, len(d.Applications))
	for _, app := range d.Applications {
		standard, standardDiags := toValues(ctx, app.StandardConfiguration)
		secured, securedDiags := toValues(ctx, app.SecuredConfiguration)
		diags.Append(standardDiags...)
		diags.Append(securedDiags...)
		out = append(out, connect.DeploymentConfigurationApplication{
			ApplicationName:       app.Name.ValueString(),
			StandardConfiguration: standard,
			SecuredConfiguration:  secured,
			Schedule:              utils.OptionalString(app.Schedule),
		})
	}
	return out, diags
}

func (d Deployment) globalConfiguration(ctx context.Context) (*connect.DeploymentGlobalConfiguration, diag.Diagnostics) {
	if d.GlobalStandardConfiguration.IsNull() && d.GlobalSecuredConfiguration.IsNull() {
		return nil, nil
	}
	standard, diags := toValues(ctx, d.GlobalStandardConfiguration)
	secured, securedDiags := toValues(ctx, d.GlobalSecuredConfiguration)
	diags.Append(securedDiags...)
	return &connect.DeploymentGlobalConfiguration{
		StandardConfiguration: standard,
		SecuredConfiguration:  secured,
	}, diags
}

// NewDeploymentFromNative fills the model from the API's answer, in the shape of prior,
// which is the plan or the state. Secured values are never returned, so they
// come from prior; so do the connector reference and the timeout, which
// describe the configuration rather than the deployment. A schedule or an
// empty configuration the configuration leaves out stays left out.
func NewDeploymentFromNative(ctx context.Context, d *connect.Deployment, prior Deployment) (Deployment, diag.Diagnostics) {
	var diags diag.Diagnostics
	current := prior
	current.ID = types.StringValue(d.ID)
	current.Key = utils.FromOptionalString(d.Key)
	current.Version = types.Int64Value(int64(d.Version))
	current.Type = types.StringValue(d.Type)
	current.Region = types.StringValue(d.DeployedRegion)
	current.Status = types.StringValue(d.Status)
	current.ConnectorVersion = types.Int64Value(int64(d.Connector.Version))
	if current.ConnectorStaged.IsNull() || current.ConnectorStaged.IsUnknown() {
		current.ConnectorStaged = types.BoolValue(d.Preview)
	}
	// After an import there is no configured reference to keep.
	imported := prior.ConnectorKey.IsNull() && prior.ConnectorID.IsNull()
	if imported {
		if d.Connector.Key != nil {
			current.ConnectorKey = types.StringValue(*d.Connector.Key)
		} else {
			current.ConnectorID = types.StringValue(d.Connector.ID)
		}
	}

	if d.GlobalConfiguration != nil {
		current.GlobalStandardConfiguration, diags = fromValues(ctx, d.GlobalConfiguration.StandardConfiguration, prior.GlobalStandardConfiguration)
	}

	apps, appDiags := alignApplications(ctx, d.Applications, prior.Applications, imported)
	diags.Append(appDiags...)
	current.Applications = apps
	return current, diags
}

// alignApplications lists the deployed applications in the order prior has
// them. An application deployed outside of the configuration, or every one
// after an import, is added in the API's order; a configuration that leaves
// one out plans a difference.
func alignApplications(
	ctx context.Context,
	deployed []connect.ApplicationDeployment,
	prior []Application,
	imported bool,
) ([]Application, diag.Diagnostics) {
	var diags diag.Diagnostics
	byName := map[string]connect.ApplicationDeployment{}
	for _, app := range deployed {
		byName[app.ApplicationName] = app
	}
	seen := map[string]bool{}
	apps := make([]Application, 0, len(deployed))
	add := func(prior Application, app connect.ApplicationDeployment) {
		current, d := alignApplication(ctx, app, prior, imported)
		diags.Append(d...)
		apps = append(apps, current)
		seen[app.ApplicationName] = true
	}
	for _, app := range prior {
		if native, ok := byName[app.Name.ValueString()]; ok {
			add(app, native)
		}
	}
	for _, app := range deployed {
		if !seen[app.ApplicationName] {
			add(Application{
				StandardConfiguration: types.MapNull(types.StringType),
				SecuredConfiguration:  types.MapNull(types.StringType),
				Schedule:              types.StringNull(),
			}, app)
		}
	}
	return apps, diags
}

// alignApplication fills one application from the API's answer. Its secured
// values come from prior, and so does a schedule the configuration leaves to
// connect.yaml, except after an import.
func alignApplication(
	ctx context.Context,
	app connect.ApplicationDeployment,
	prior Application,
	imported bool,
) (Application, diag.Diagnostics) {
	standard, diags := fromValues(ctx, app.StandardConfiguration, prior.StandardConfiguration)
	schedule := prior.Schedule
	if !schedule.IsNull() || imported {
		schedule = utils.FromOptionalString(app.Schedule)
	}
	return Application{
		Name:                  types.StringValue(app.ApplicationName),
		StandardConfiguration: standard,
		SecuredConfiguration:  prior.SecuredConfiguration,
		Schedule:              schedule,
		URL:                   utils.FromOptionalString(app.Url),
		Topic:                 utils.FromOptionalString(app.Topic),
	}, diags
}

func toValues(ctx context.Context, m types.Map) ([]connect.ConfigurationValue, diag.Diagnostics) {
	out := []connect.ConfigurationValue{}
	if m.IsNull() || m.IsUnknown() {
		return out, nil
	}
	values := map[string]string{}
	diags := m.ElementsAs(ctx, &values, false)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out = append(out, connect.ConfigurationValue{Key: key, Value: values[key]})
	}
	return out, diags
}

// fromValues keeps prior's null when the API answers with no values, so an
// application without standard configuration plans no difference.
func fromValues(ctx context.Context, values []connect.ConfigurationValue, prior types.Map) (types.Map, diag.Diagnostics) {
	if len(values) == 0 && (prior.IsNull() || prior.IsUnknown()) {
		return types.MapNull(types.StringType), nil
	}
	m := map[string]string{}
	for _, value := range values {
		m[value.Key] = value.Value
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

func optionalBool(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return utils.Ref(value.ValueBool())
}
