package mcp_server

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

const CommerceMCP = "CommerceMCP"

// McpServer is the main resource schema data
type McpServer struct {
	ID                   types.String                     `tfsdk:"id"`
	Version              types.Int64                      `tfsdk:"version"`
	Key                  types.String                     `tfsdk:"key"`
	Name                 customtypes.LocalizedStringValue `tfsdk:"name"`
	Description          customtypes.LocalizedStringValue `tfsdk:"description"`
	AuthenticationMode   types.String                     `tfsdk:"authentication_mode"`
	State                types.String                     `tfsdk:"state"`
	Type                 types.String                     `tfsdk:"type"`
	MajorVersion         types.String                     `tfsdk:"major_version"`
	URL                  types.String                     `tfsdk:"url"`
	Tools                []types.String                   `tfsdk:"tools"`
	ToolCustomizations   []ToolCustomization              `tfsdk:"tool_customization"`
	JsonOutputFiltering  []JsonOutputFiltering            `tfsdk:"json_output_filtering"`
	ToolOutputFormatting types.String                     `tfsdk:"tool_output_formatting"`
}

type ToolCustomization struct {
	Tool        types.String        `tfsdk:"tool"`
	Description types.String        `tfsdk:"description"`
	Parameters  []ParameterOverride `tfsdk:"parameter"`
}

// ParameterOverride carries the API's free-form values as JSON, as
// jsonencode() writes them.
type ParameterOverride struct {
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ConstJSON        types.String `tfsdk:"const_json"`
	DefaultJSON      types.String `tfsdk:"default_json"`
	AllowedTypesJSON types.String `tfsdk:"allowed_types_json"`
}

type JsonOutputFiltering struct {
	Redact           []Matcher      `tfsdk:"redact"`
	Drop             []Matcher      `tfsdk:"drop"`
	FieldPathsToKeep []types.String `tfsdk:"field_paths_to_keep"`
	RedactedWith     types.String   `tfsdk:"redacted_with"`
}

type Matcher struct {
	FieldPaths        []types.String `tfsdk:"field_paths"`
	FieldNames        []types.String `tfsdk:"field_names"`
	FieldNameContains []types.String `tfsdk:"field_name_contains"`
}

func (m McpServer) draft() (platform.McpServerDraft, error) {
	customizations, err := toCustomizations(m.ToolCustomizations)
	if err != nil {
		return platform.McpServerDraft{}, err
	}
	authentication := platform.McpServerAuthenticationMode(m.AuthenticationMode.ValueString())
	state := platform.McpServerState(m.State.ValueString())
	return platform.McpServerDraft{
		Key:                m.Key.ValueString(),
		Name:               m.Name.ValueLocalizedStringRef(),
		Description:        m.Description.ValueLocalizedStringRef(),
		AuthenticationMode: &authentication,
		State:              &state,
		McpServer: platform.CommerceMcpServerConfigDraft{
			MajorVersion:         m.MajorVersion.ValueString(),
			Tools:                toTools(m.Tools),
			ToolCustomizations:   customizations,
			JsonOutputFiltering:  toFiltering(m.JsonOutputFiltering),
			ToolOutputFormatting: toFormatting(m.ToolOutputFormatting),
		},
	}, nil
}

func (m McpServer) updateActions(plan McpServer) (platform.McpServerUpdate, error) {
	update := platform.McpServerUpdate{
		Version: int(m.Version.ValueInt64()),
		Actions: []platform.McpServerUpdateAction{},
	}
	if !reflect.DeepEqual(m.Name, plan.Name) {
		update.Actions = append(update.Actions, platform.McpServerSetNameAction{Name: plan.Name.ValueLocalizedStringRef()})
	}
	if !reflect.DeepEqual(m.Description, plan.Description) {
		update.Actions = append(update.Actions, platform.McpServerSetDescriptionAction{
			Description: plan.Description.ValueLocalizedStringRef(),
		})
	}
	if !m.AuthenticationMode.Equal(plan.AuthenticationMode) {
		update.Actions = append(update.Actions, platform.McpServerSetAuthenticationModeAction{
			AuthenticationMode: platform.McpServerAuthenticationMode(plan.AuthenticationMode.ValueString()),
		})
	}
	if !m.State.Equal(plan.State) {
		update.Actions = append(update.Actions, platform.McpServerSetStateAction{
			State: platform.McpServerState(plan.State.ValueString()),
		})
	}
	if !reflect.DeepEqual(sortedStrings(m.Tools), sortedStrings(plan.Tools)) {
		update.Actions = append(update.Actions, platform.McpServerSetToolsAction{Tools: toTools(plan.Tools)})
	}
	if !reflect.DeepEqual(m.ToolCustomizations, plan.ToolCustomizations) {
		customizations, err := toCustomizations(plan.ToolCustomizations)
		if err != nil {
			return update, err
		}
		update.Actions = append(update.Actions, platform.McpServerSetToolCustomizationsAction{
			ToolCustomizations: customizations,
		})
	}
	if !reflect.DeepEqual(m.JsonOutputFiltering, plan.JsonOutputFiltering) {
		update.Actions = append(update.Actions, platform.McpServerSetJsonOutputFilteringAction{
			JsonOutputFiltering: toFiltering(plan.JsonOutputFiltering),
		})
	}
	if !m.ToolOutputFormatting.Equal(plan.ToolOutputFormatting) {
		update.Actions = append(update.Actions, platform.McpServerSetToolOutputFormattingAction{
			ToolOutputFormatting: toFormatting(plan.ToolOutputFormatting),
		})
	}
	return update, nil
}

// NewMcpServerFromNative maps the API's MCP Server, keeping prior's JSON
// strings where they encode the same values, so a different key order or
// spacing plans no difference, and prior's order of tools.
func NewMcpServerFromNative(n *platform.McpServer, prior McpServer) McpServer {
	current := McpServer{
		ID:                   types.StringValue(n.ID),
		Version:              types.Int64Value(int64(n.Version)),
		Key:                  types.StringValue(n.Key),
		Name:                 utils.FromOptionalLocalizedString(n.Name),
		Description:          utils.FromOptionalLocalizedString(n.Description),
		AuthenticationMode:   types.StringValue(string(n.AuthenticationMode)),
		State:                types.StringValue(string(n.State)),
		Type:                 types.StringValue(CommerceMCP),
		MajorVersion:         types.StringNull(),
		URL:                  types.StringNull(),
		Tools:                []types.String{},
		ToolCustomizations:   nil,
		JsonOutputFiltering:  nil,
		ToolOutputFormatting: types.StringNull(),
	}
	config, ok := commerceConfig(n.McpServer)
	if !ok {
		return current
	}
	current.MajorVersion = types.StringValue(config.MajorVersion)
	current.URL = types.StringValue(config.Url)
	current.Tools = fromTools(config.Tools, prior.Tools)
	current.ToolCustomizations = fromCustomizations(config.ToolCustomizations, prior.ToolCustomizations)
	current.JsonOutputFiltering = fromFiltering(config.JsonOutputFiltering)
	if config.ToolOutputFormatting != nil {
		current.ToolOutputFormatting = types.StringValue(string(*config.ToolOutputFormatting))
	}
	return current
}

func commerceConfig(config platform.McpServerConfig) (platform.CommerceMcpServerConfig, bool) {
	switch c := config.(type) {
	case platform.CommerceMcpServerConfig:
		return c, true
	case *platform.CommerceMcpServerConfig:
		return *c, true
	}
	return platform.CommerceMcpServerConfig{}, false
}

func toTools(tools []types.String) []platform.McpServerTool {
	out := make([]platform.McpServerTool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, platform.McpServerTool(tool.ValueString()))
	}
	return out
}

// fromTools lists the tools in prior's order when they are the same set.
func fromTools(tools []platform.McpServerTool, prior []types.String) []types.String {
	out := make([]types.String, 0, len(tools))
	for _, tool := range tools {
		out = append(out, types.StringValue(string(tool)))
	}
	if reflect.DeepEqual(sortedStrings(out), sortedStrings(prior)) {
		return prior
	}
	return out
}

func sortedStrings(values []types.String) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, v.ValueString())
	}
	sort.Strings(out)
	return out
}

func toCustomizations(in []ToolCustomization) ([]platform.ToolCustomization, error) {
	out := make([]platform.ToolCustomization, 0, len(in))
	for _, c := range in {
		parameters := make([]platform.ParameterOverride, 0, len(c.Parameters))
		for _, p := range c.Parameters {
			parameter := platform.ParameterOverride{
				Name:        p.Name.ValueString(),
				Description: utils.OptionalString(p.Description),
			}
			var err error
			if parameter.Const, err = decodeJSON(p.ConstJSON); err != nil {
				return nil, err
			}
			if parameter.Default, err = decodeJSON(p.DefaultJSON); err != nil {
				return nil, err
			}
			allowed, err := decodeJSON(p.AllowedTypesJSON)
			if err != nil {
				return nil, err
			}
			if list, ok := allowed.([]any); ok {
				parameter.AllowedTypes = list
			}
			parameters = append(parameters, parameter)
		}
		out = append(out, platform.ToolCustomization{
			Tool:        platform.McpServerTool(c.Tool.ValueString()),
			Description: utils.OptionalString(c.Description),
			Parameters:  parameters,
		})
	}
	return out, nil
}

func fromCustomizations(in []platform.ToolCustomization, prior []ToolCustomization) []ToolCustomization {
	if len(in) == 0 {
		return nil
	}
	priorByTool := map[string]ToolCustomization{}
	for _, c := range prior {
		priorByTool[c.Tool.ValueString()] = c
	}
	out := make([]ToolCustomization, 0, len(in))
	for _, c := range in {
		priorParameters := map[string]ParameterOverride{}
		for _, p := range priorByTool[string(c.Tool)].Parameters {
			priorParameters[p.Name.ValueString()] = p
		}
		var parameters []ParameterOverride
		for _, p := range c.Parameters {
			prior := priorParameters[p.Name]
			var allowed any
			if p.AllowedTypes != nil {
				allowed = p.AllowedTypes
			}
			parameters = append(parameters, ParameterOverride{
				Name:             types.StringValue(p.Name),
				Description:      utils.FromOptionalString(p.Description),
				ConstJSON:        encodeJSON(p.Const, prior.ConstJSON),
				DefaultJSON:      encodeJSON(p.Default, prior.DefaultJSON),
				AllowedTypesJSON: encodeJSON(allowed, prior.AllowedTypesJSON),
			})
		}
		out = append(out, ToolCustomization{
			Tool:        types.StringValue(string(c.Tool)),
			Description: utils.FromOptionalString(c.Description),
			Parameters:  parameters,
		})
	}
	return out
}

func decodeJSON(value types.String) (any, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var out any
	err := json.Unmarshal([]byte(value.ValueString()), &out)
	return out, err
}

// encodeJSON is value as JSON, or prior when prior encodes the same value.
func encodeJSON(value any, prior types.String) types.String {
	if value == nil {
		return types.StringNull()
	}
	if decoded, err := decodeJSON(prior); err == nil && reflect.DeepEqual(normalize(decoded), normalize(value)) {
		return prior
	}
	data, err := json.Marshal(value)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(data))
}

// normalize gives a value the shape json.Unmarshal would, so values read from
// the API and from configuration compare equal.
func normalize(value any) any {
	data, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var out any
	_ = json.Unmarshal(data, &out)
	return out
}

func toMatcher(in []Matcher) *platform.McpServerJsonOutputFilteringMatcher {
	if len(in) == 0 {
		return nil
	}
	return &platform.McpServerJsonOutputFilteringMatcher{
		FieldPaths:        stringsOf(in[0].FieldPaths),
		FieldNames:        stringsOf(in[0].FieldNames),
		FieldNameContains: stringsOf(in[0].FieldNameContains),
	}
}

func fromMatcher(in *platform.McpServerJsonOutputFilteringMatcher) []Matcher {
	if in == nil {
		return nil
	}
	return []Matcher{{
		FieldPaths:        valuesOf(in.FieldPaths),
		FieldNames:        valuesOf(in.FieldNames),
		FieldNameContains: valuesOf(in.FieldNameContains),
	}}
}

func toFiltering(in []JsonOutputFiltering) *platform.McpServerJsonOutputFiltering {
	if len(in) == 0 {
		return nil
	}
	return &platform.McpServerJsonOutputFiltering{
		Redact:           toMatcher(in[0].Redact),
		Drop:             toMatcher(in[0].Drop),
		FieldPathsToKeep: stringsOf(in[0].FieldPathsToKeep),
		RedactedWith:     utils.OptionalString(in[0].RedactedWith),
	}
}

func fromFiltering(in *platform.McpServerJsonOutputFiltering) []JsonOutputFiltering {
	if in == nil {
		return nil
	}
	return []JsonOutputFiltering{{
		Redact:           fromMatcher(in.Redact),
		Drop:             fromMatcher(in.Drop),
		FieldPathsToKeep: valuesOf(in.FieldPathsToKeep),
		RedactedWith:     utils.FromOptionalString(in.RedactedWith),
	}}
}

func toFormatting(value types.String) *platform.McpServerToolOutputFormatting {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	formatting := platform.McpServerToolOutputFormatting(value.ValueString())
	return &formatting
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

// valuesOf keeps an absent list absent, so a matcher reads back the way it
// is configured.
func valuesOf(in []string) []types.String {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.String, 0, len(in))
	for _, v := range in {
		out = append(out, types.StringValue(v))
	}
	return out
}
