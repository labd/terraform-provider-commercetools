package subscription

import (
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Subscription is the main resource schema data
type Subscription struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Version     types.Int64  `tfsdk:"version"`
	Destination *Destination `tfsdk:"destination"`
	Format      *Format      `tfsdk:"format"`
	Messages    []Message    `tfsdk:"message"`
	Changes     []Changes    `tfsdk:"changes"`
}

type Changes struct {
	ResourceTypeIds []types.String `tfsdk:"resource_type_ids"`
}

func (c Changes) ToNative() []platform.ChangeSubscription {
	result := make([]platform.ChangeSubscription, len(c.ResourceTypeIds))

	for i := range c.ResourceTypeIds {
		val := c.ResourceTypeIds[i].ValueString()
		result[i] = platform.ChangeSubscription{
			ResourceTypeId: platform.ChangeSubscriptionResourceTypeId(val),
		}
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

func (d *Destination) Import(n platform.Destination, state *Destination) {
	switch v := n.(type) {
	case platform.AzureEventGridDestination:
		d.Type = types.StringValue("EventGrid")
		d.URI = types.StringValue(v.Uri)

		if state != nil {
			d.AccessKey = state.AccessKey
		} else {
			d.AccessKey = types.StringUnknown()
		}
	case platform.AzureServiceBusDestination:
		d.Type = types.StringValue("AzureServiceBus")
		d.ConnectionString = types.StringValue(v.ConnectionString)
	case platform.EventBridgeDestination:
		d.Type = types.StringValue("EventBridge")
		d.AccountID = types.StringValue(v.AccountId)
		d.Region = types.StringValue(v.Region)
	case platform.GoogleCloudPubSubDestination:
		d.Type = types.StringValue("GoogleCloudPubSub")
		d.ProjectID = types.StringValue(v.ProjectId)
		d.Topic = types.StringValue(v.Topic)
	case platform.SnsDestination:
		d.Type = types.StringValue("SNS")
		d.TopicARN = types.StringValue(v.TopicArn)
		d.AccessKey = utils.FromOptionalString(v.AccessKey)

		if state != nil {
			d.AccessSecret = state.AccessSecret
		} else {
			d.AccessSecret = types.StringUnknown()
		}
	case platform.SqsDestination:
		d.Type = types.StringValue("SQS")
		d.QueueURL = types.StringValue(v.QueueUrl)
		d.AccessKey = utils.FromOptionalString(v.AccessKey)
		d.Region = types.StringValue(v.Region)

		if state != nil {
			d.AccessSecret = state.AccessSecret
		} else {
			d.AccessSecret = types.StringUnknown()
		}
	}
}

func (d Destination) ToNative() platform.Destination {
	val := d.Type.ValueString()

	switch val {
	case "SQS":
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
	case "SNS":
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

type Format struct {
	Type              types.String `tfsdk:"type"`
	CloudEventVersion types.String `tfsdk:"cloud_events_version"`
}

func (f Format) ToNative() platform.DeliveryFormat {
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

func (f *Format) Import(n platform.DeliveryFormat) {
	switch v := n.(type) {
	case platform.PlatformFormat:
		f.Type = types.StringValue("Platform")
	case platform.CloudEventsFormat:
		f.Type = types.StringValue("CloudEvents")
		f.CloudEventVersion = types.StringValue(v.CloudEventsVersion)
	}
}

type Message struct {
	ResourceTypeID types.String   `tfsdk:"resource_type_id"`
	Types          []types.String `tfsdk:"types"`
}

func (m Message) ToNative() platform.MessageSubscription {
	return platform.MessageSubscription{
		ResourceTypeId: platform.MessageSubscriptionResourceTypeId(m.ResourceTypeID.ValueString()),
		Types: pie.Map(m.Types, func(v types.String) string {
			return v.ValueString()
		}),
	}
}

func (s Subscription) Draft() platform.SubscriptionDraft {
	changes := []platform.ChangeSubscription{}
	for _, c := range s.Changes {
		changes = append(changes, c.ToNative()...)
	}

	draft := platform.SubscriptionDraft{
		Key:         utils.OptionalString(s.Key),
		Destination: s.Destination.ToNative(),
		Messages: pie.Map(s.Messages, func(m Message) platform.MessageSubscription {
			return m.ToNative()
		}),
		Changes: changes,
	}

	if s.Format != nil {
		draft.Format = s.Format.ToNative()
	}

	return draft
}

func (s *Subscription) Import(n *platform.Subscription, state Subscription) {
	s.ID = types.StringValue(n.ID)
	s.Version = types.Int64Value(int64(n.Version))
	s.Key = utils.FromOptionalString(n.Key)

	if s.Format == nil {
		s.Format = &Format{}
	}
	s.Format.Import(n.Format)

	if s.Destination == nil {
		s.Destination = &Destination{}
	}
	s.Destination.Import(n.Destination, state.Destination)

	s.Changes = []Changes{
		{
			ResourceTypeIds: pie.Map(n.Changes, func(c platform.ChangeSubscription) types.String {
				return types.StringValue(string(c.ResourceTypeId))
			}),
		},
	}

	s.Messages = make([]Message, len(n.Messages))
	for i, message := range n.Messages {
		s.Messages[i] = Message{
			ResourceTypeID: types.StringValue(string(message.ResourceTypeId)),
			Types:          pie.Map(message.Types, types.StringValue),
		}
	}
}

func (s Subscription) UpdateActions(n Subscription) platform.SubscriptionUpdate {
	result := platform.SubscriptionUpdate{
		Version: int(s.Version.ValueInt64()),
		Actions: []platform.SubscriptionUpdateAction{},
	}

	if s.Key != n.Key {
		var value *string
		if !n.Key.IsNull() && !n.Key.IsUnknown() {
			value = utils.StringRef(n.Key.ValueString())
		}
		result.Actions = append(
			result.Actions,
			platform.SubscriptionSetKeyAction{Key: value})
	}

	if !reflect.DeepEqual(s.Destination, n.Destination) {
		result.Actions = append(
			result.Actions,
			platform.SubscriptionChangeDestinationAction{Destination: n.Destination.ToNative()})
	}

	if !reflect.DeepEqual(s.Messages, n.Messages) {
		messages := pie.Map(n.Messages, func(m Message) platform.MessageSubscription {
			return m.ToNative()
		})
		result.Actions = append(
			result.Actions,
			platform.SubscriptionSetMessagesAction{Messages: messages})
	}

	return result
}
