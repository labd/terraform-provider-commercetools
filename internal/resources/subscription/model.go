package subscription

import (
	"reflect"
	"regexp"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

const (
	AzureServiceBus   = "AzureServiceBus"
	EventBridge       = "EventBridge"
	EventGrid         = "EventGrid"
	GoogleCloudPubSub = "GoogleCloudPubSub"
	SNS               = "SNS"
	SQS               = "SQS"
)

// Subscription is the main resource schema data
type Subscription struct {
	ID          types.String  `tfsdk:"id"`
	Key         types.String  `tfsdk:"key"`
	Version     types.Int64   `tfsdk:"version"`
	Destination []Destination `tfsdk:"destination"`
	Format      []Format      `tfsdk:"format"`
	Messages    []Message     `tfsdk:"message"`
	Changes     []Changes     `tfsdk:"changes"`
}

func NewSubscriptionFromNative(n *platform.Subscription) Subscription {
	res := Subscription{
		ID:          types.StringValue(n.ID),
		Version:     types.Int64Value(int64(n.Version)),
		Key:         utils.FromOptionalString(n.Key),
		Format:      []Format{},
		Destination: []Destination{},
		Messages:    make([]Message, len(n.Messages)),
		Changes:     []Changes{},
	}

	format := NewFormatFromNative(n.Format)
	res.Format = append(res.Format, *format)

	dst := NewDestinationFromNative(n.Destination)
	res.Destination = append(res.Destination, *dst)

	if len(n.Changes) > 0 {
		item := Changes{
			ResourceTypeIds: pie.Map(n.Changes, func(c platform.ChangeSubscription) types.String {
				return types.StringValue(string(c.ResourceTypeId))
			}),
		}
		res.Changes = append(res.Changes, item)
	}

	for i, message := range n.Messages {
		res.Messages[i] = Message{
			ResourceTypeID: types.StringValue(string(message.ResourceTypeId)),
			Types:          pie.Map(message.Types, types.StringValue),
		}
	}

	return res
}

func (s *Subscription) matchDefaults(state Subscription) {
	if len(state.Format) == 0 {
		if len(s.Format) == 1 && s.Format[0].Type.ValueString() == "Platform" {
			s.Format = []Format{}
		}
	}
}

func (s *Subscription) setSecretValues(state Subscription) {
	s.Destination[0].setSecretValues(&state.Destination[0])
}

func (s *Subscription) draft() platform.SubscriptionDraft {
	var changes []platform.ChangeSubscription
	for _, c := range s.Changes {
		changes = append(changes, c.toNative()...)
	}

	draft := platform.SubscriptionDraft{
		Key:         utils.OptionalString(s.Key),
		Destination: s.Destination[0].ToNative(),
		Messages: pie.Map(s.Messages, func(m Message) platform.MessageSubscription {
			return m.toNative()
		}),
		Changes: changes,
	}

	if len(s.Format) > 0 {
		draft.Format = s.Format[0].toNative()
	}

	return draft
}

func (s *Subscription) updateActions(plan Subscription) platform.SubscriptionUpdate {
	result := platform.SubscriptionUpdate{
		Version: int(s.Version.ValueInt64()),
		Actions: []platform.SubscriptionUpdateAction{},
	}

	// setKey
	if s.Key != plan.Key {
		var value *string
		if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
			value = utils.StringRef(plan.Key.ValueString())
		}
		result.Actions = append(
			result.Actions,
			platform.SubscriptionSetKeyAction{Key: value})
	}

	// changeDestination
	if !reflect.DeepEqual(s.Destination, plan.Destination) {
		result.Actions = append(
			result.Actions,
			platform.SubscriptionChangeDestinationAction{
				Destination: plan.Destination[0].ToNative(),
			})
	}

	// setChanges
	if !reflect.DeepEqual(s.Changes, plan.Changes) {
		var changes []platform.ChangeSubscription
		for _, c := range plan.Changes {
			changes = append(changes, c.toNative()...)
		}

		result.Actions = append(
			result.Actions,
			platform.SubscriptionSetChangesAction{
				Changes: changes,
			})
	}

	// setMessages
	if !reflect.DeepEqual(s.Messages, plan.Messages) {
		messages := pie.Map(plan.Messages, func(m Message) platform.MessageSubscription {
			return m.toNative()
		})
		result.Actions = append(
			result.Actions,
			platform.SubscriptionSetMessagesAction{Messages: messages})
	}

	return result
}

type Destination struct {
	Type     types.String `tfsdk:"type"`
	TopicARN types.String `tfsdk:"topic_arn"`

	// SNS, SQS, EventGrid
	AccessKey types.String `tfsdk:"access_key"`

	// SNS, SQS, EventGrid
	AccessSecret types.String `tfsdk:"access_secret"`

	// SQS
	QueueURL types.String `tfsdk:"queue_url"`

	AccountID types.String `tfsdk:"account_id"`

	// SQS, SNS
	Region types.String `tfsdk:"region"`

	// EventGrid
	URI types.String `tfsdk:"uri"`

	// AzureServiceBus
	ConnectionString types.String `tfsdk:"connection_string"`

	// For GooglePubSub
	ProjectID types.String `tfsdk:"project_id"`
	Topic     types.String `tfsdk:"topic"`
}

func (d *Destination) setSecretValues(state *Destination) {
	if state == nil {
		return
	}

	switch d.Type.ValueString() {
	case AzureServiceBus:
		// Quick hack. Filter out the shared access key since that value is
		// masked by commercetools. If the strings are equal then copy the val
		// from the state. Otherwise, we use the value from the plan
		re := regexp.MustCompile(`;?SharedAccessKey=[^;]+`)
		planVal := re.ReplaceAllString(d.ConnectionString.ValueString(), "")
		stateVal := re.ReplaceAllString(state.ConnectionString.ValueString(), "")
		if planVal == stateVal {
			d.ConnectionString = state.ConnectionString
		}
	case EventGrid:
		if d.AccessKey.IsUnknown() {
			d.AccessKey = state.AccessKey
		}
	case SNS, SQS:
		if d.AccessSecret.IsNull() {
			d.AccessSecret = state.AccessSecret
		}
	}
}

func NewDestinationFromNative(n platform.Destination) *Destination {
	d := &Destination{}
	switch v := n.(type) {
	case platform.AzureEventGridDestination:
		d.Type = types.StringValue(EventGrid)
		d.URI = types.StringValue(v.Uri)
		d.AccessKey = types.StringUnknown()
	case platform.AzureServiceBusDestination:
		d.Type = types.StringValue(AzureServiceBus)
		d.ConnectionString = types.StringValue(v.ConnectionString)
	case platform.EventBridgeDestination:
		d.Type = types.StringValue(EventBridge)
		d.AccountID = types.StringValue(v.AccountId)
		d.Region = types.StringValue(v.Region)
	case platform.GoogleCloudPubSubDestination:
		d.Type = types.StringValue(GoogleCloudPubSub)
		d.ProjectID = types.StringValue(v.ProjectId)
		d.Topic = types.StringValue(v.Topic)
	case platform.SnsDestination:
		d.Type = types.StringValue(SNS)
		d.TopicARN = types.StringValue(v.TopicArn)
		d.AccessKey = utils.FromOptionalString(v.AccessKey)
		d.AccessSecret = types.StringNull()
	case platform.SqsDestination:
		d.Type = types.StringValue(SQS)
		d.QueueURL = types.StringValue(v.QueueUrl)
		d.Region = types.StringValue(v.Region)
		d.AccessKey = utils.FromOptionalString(v.AccessKey)
		d.AccessSecret = types.StringNull()
	}
	return d
}

func (d *Destination) ToNative() platform.Destination {
	val := d.Type.ValueString()

	switch val {
	case AzureServiceBus:
		return platform.AzureServiceBusDestination{
			ConnectionString: d.ConnectionString.ValueString(),
		}
	case EventBridge:
		return platform.EventBridgeDestination{
			Region:    d.Region.ValueString(),
			AccountId: d.AccountID.ValueString(),
		}
	case EventGrid:
		return platform.AzureEventGridDestination{
			AccessKey: d.AccessKey.ValueString(),
			Uri:       d.URI.ValueString(),
		}
	case GoogleCloudPubSub:
		return platform.GoogleCloudPubSubDestination{
			ProjectId: d.ProjectID.ValueString(),
			Topic:     d.Topic.ValueString(),
		}
	case SQS:
		result := platform.SqsDestination{
			AccessKey:    utils.OptionalString(d.AccessKey),
			AccessSecret: utils.OptionalString(d.AccessSecret),
			QueueUrl:     d.QueueURL.ValueString(),
			Region:       d.Region.ValueString(),
		}
		if result.AccessKey == nil {
			authMode := platform.AwsAuthenticationModeIAM
			result.AuthenticationMode = &authMode
		}
		return result
	case SNS:
		result := platform.SnsDestination{
			AccessKey:    utils.OptionalString(d.AccessKey),
			AccessSecret: utils.OptionalString(d.AccessSecret),
			TopicArn:     d.TopicARN.ValueString(),
		}
		if result.AccessKey == nil {
			authMode := platform.AwsAuthenticationModeIAM
			result.AuthenticationMode = &authMode
		}
		return result
	}
	return nil
}

type Changes struct {
	ResourceTypeIds []types.String `tfsdk:"resource_type_ids"`
}

func (c Changes) toNative() []platform.ChangeSubscription {
	result := make([]platform.ChangeSubscription, len(c.ResourceTypeIds))

	for i := range c.ResourceTypeIds {
		val := c.ResourceTypeIds[i].ValueString()
		result[i] = platform.ChangeSubscription{
			ResourceTypeId: platform.ChangeSubscriptionResourceTypeId(val),
		}
	}
	return result
}

type Format struct {
	Type              types.String `tfsdk:"type"`
	CloudEventVersion types.String `tfsdk:"cloud_events_version"`
}

func (f Format) toNative() platform.DeliveryFormat {
	if f.Type.IsUnknown() || f.Type.IsNull() {
		return nil
	}

	val := f.Type.ValueString()
	switch val {
	case "Platform":
		return platform.PlatformFormat{}
	case "CloudEvents":
		version := "1.0"
		if !f.CloudEventVersion.IsNull() {
			version = f.CloudEventVersion.ValueString()
		}
		return platform.CloudEventsFormat{
			CloudEventsVersion: version,
		}
	}
	return nil
}

func NewFormatFromNative(n platform.DeliveryFormat) *Format {
	switch v := n.(type) {
	case platform.CloudEventsFormat:
		return &Format{
			Type:              types.StringValue("CloudEvents"),
			CloudEventVersion: types.StringValue(v.CloudEventsVersion),
		}
	default:
		return &Format{
			Type:              types.StringValue("Platform"),
			CloudEventVersion: types.StringNull(),
		}
	}
}

type Message struct {
	ResourceTypeID types.String   `tfsdk:"resource_type_id"`
	Types          []types.String `tfsdk:"types"`
}

func (m Message) toNative() platform.MessageSubscription {
	return platform.MessageSubscription{
		ResourceTypeId: platform.MessageSubscriptionResourceTypeId(m.ResourceTypeID.ValueString()),
		Types: pie.Map(m.Types, func(v types.String) string {
			return v.ValueString()
		}),
	}
}
