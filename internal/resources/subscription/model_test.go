package subscription

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestDestination(t *testing.T) {
	native := platform.SqsDestination{
		QueueUrl: "https://sqs.eu-central-1.amazonaws.com/123456789012/terraform-test",
		Region:   "eu-central-1",
	}
	dest := NewDestinationFromNative(native)
	assert.Equal(t, dest, Destination{
		Type:         types.StringValue("SQS"),
		QueueURL:     types.StringValue("https://sqs.eu-central-1.amazonaws.com/123456789012/terraform-test"),
		Region:       types.StringValue("eu-central-1"),
		AccessKey:    types.StringNull(),
		AccessSecret: types.StringNull(),
	})

}

func TestImport(t *testing.T) {
	testCases := []struct {
		name     string
		n        platform.Destination
		state    *Destination
		wantDest Destination
	}{
		{
			name: "AzureEventGridDestination",
			n: platform.AzureEventGridDestination{
				Uri: "https://example.com",
			},
			state: nil,
			wantDest: Destination{
				Type:      types.StringValue("EventGrid"),
				URI:       types.StringValue("https://example.com"),
				AccessKey: types.StringUnknown(),
			},
		},
		{
			name: "AzureEventGridDestination (state)",
			n: platform.AzureEventGridDestination{
				Uri: "https://example.com",
			},
			state: &Destination{
				AccessKey: types.StringValue("test-key"),
			},
			wantDest: Destination{
				Type:      types.StringValue("EventGrid"),
				URI:       types.StringValue("https://example.com"),
				AccessKey: types.StringValue("test-key"),
			},
		},
		{
			name: "AzureServiceBusDestination",
			n: platform.AzureServiceBusDestination{
				ConnectionString: "test-connection-string",
			},
			wantDest: Destination{
				Type:             types.StringValue("AzureServiceBus"),
				ConnectionString: types.StringValue("test-connection-string"),
			},
		},
		{
			name: "AzureServiceBusDestination (state)",
			n: platform.AzureServiceBusDestination{
				ConnectionString: "Endpoint=sb://michael-temp.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=****1Fw=;EntityPath=my-test-queue",
			},
			state: &Destination{
				Type:             types.StringValue("AzureServiceBus"),
				ConnectionString: types.StringValue("Endpoint=sb://michael-temp.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=17108y4812311Fw=;EntityPath=my-test-queue"),
			},
			wantDest: Destination{
				Type:             types.StringValue("AzureServiceBus"),
				ConnectionString: types.StringValue("Endpoint=sb://michael-temp.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=17108y4812311Fw=;EntityPath=my-test-queue"),
			},
		},
		{
			name: "EventBridgeDestination",
			n: platform.EventBridgeDestination{
				AccountId: "test-account-id",
				Region:    "test-region",
			},
			wantDest: Destination{
				Type:      types.StringValue("EventBridge"),
				AccountID: types.StringValue("test-account-id"),
				Region:    types.StringValue("test-region"),
			},
		},
		{
			name: "GoogleCloudPubSubDestination",
			n: platform.GoogleCloudPubSubDestination{
				ProjectId: "test-project-id",
				Topic:     "test-topic",
			},
			wantDest: Destination{
				Type:      types.StringValue("GoogleCloudPubSub"),
				ProjectID: types.StringValue("test-project-id"),
				Topic:     types.StringValue("test-topic"),
			},
		},
		{
			name: "SnsDestination",
			n: platform.SnsDestination{
				TopicArn: "test-topic-arn",
			},
			state: nil,
			wantDest: Destination{
				Type:         types.StringValue("SNS"),
				TopicARN:     types.StringValue("test-topic-arn"),
				AccessKey:    types.StringNull(),
				AccessSecret: types.StringNull(),
			},
		},
		{
			name: "SnsDestination (state)",
			n: platform.SnsDestination{
				TopicArn:  "test-topic-arn",
				AccessKey: utils.StringRef("foobar"),
			},
			state: &Destination{
				AccessSecret: types.StringValue("secret"),
			},
			wantDest: Destination{
				Type:         types.StringValue("SNS"),
				TopicARN:     types.StringValue("test-topic-arn"),
				AccessKey:    types.StringValue("foobar"),
				AccessSecret: types.StringValue("secret"),
			},
		},
		{
			name: "SqsDestination",
			n: platform.SqsDestination{
				QueueUrl: "test-queue-url",
				Region:   "test-region",
			},
			state: nil,
			wantDest: Destination{
				Type:         types.StringValue("SQS"),
				Region:       types.StringValue("test-region"),
				QueueURL:     types.StringValue("test-queue-url"),
				AccessKey:    types.StringNull(),
				AccessSecret: types.StringNull(),
			},
		},
		{
			name: "SqsDestination (state)",
			n: platform.SqsDestination{
				QueueUrl:  "test-queue-url",
				Region:    "test-region",
				AccessKey: utils.StringRef("foobar"),
			},
			state: &Destination{
				AccessSecret: types.StringValue("secret"),
			},
			wantDest: Destination{
				Type:         types.StringValue("SQS"),
				Region:       types.StringValue("test-region"),
				QueueURL:     types.StringValue("test-queue-url"),
				AccessKey:    types.StringValue("foobar"),
				AccessSecret: types.StringValue("secret"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDestinationFromNative(tc.n)
			d.SetStateData(tc.state)
			assert.EqualValues(t, tc.wantDest, *d)
		})
	}
}

func TestUpdateActions(t *testing.T) {
	testCases := []struct {
		name     string
		state    Subscription
		plan     Subscription
		expected platform.SubscriptionUpdate
	}{
		{
			name: "test",
			state: Subscription{
				Version: types.Int64Value(10),
				Key:     types.StringValue("foo"),
			},
			plan: Subscription{
				Key: types.StringValue("foobar"),
			},
			expected: platform.SubscriptionUpdate{
				Version: 10,
				Actions: []platform.SubscriptionUpdateAction{
					platform.SubscriptionSetKeyAction{
						Key: utils.StringRef("foobar"),
					},
				},
			},
		},
		{
			name: "test",
			state: Subscription{
				Version: types.Int64Value(10),
				Key:     types.StringValue("foo"),
			},
			plan: Subscription{
				Key: types.StringNull(),
				Messages: []Message{
					{
						ResourceTypeID: types.StringValue("product"),
					},
				},
			},
			expected: platform.SubscriptionUpdate{
				Version: 10,
				Actions: []platform.SubscriptionUpdateAction{
					platform.SubscriptionSetKeyAction{
						Key: nil,
					},
					platform.SubscriptionSetMessagesAction{
						Messages: []platform.MessageSubscription{
							{
								ResourceTypeId: "product",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.state.UpdateActions(tc.plan)
			assert.EqualValues(t, tc.expected, result)
		})
	}
}
