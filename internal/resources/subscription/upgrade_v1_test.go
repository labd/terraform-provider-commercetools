package subscription

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_upgradeStateV1(t *testing.T) {
	oldState := []byte(`
	  {
		"changes": [
		  {
			"resource_type_ids": [
			  "product"
			]
		  }
		],
		"destination": [
		  {
			"access_key": "",
			"access_secret": "",
			"account_id": "",
			"connection_string": "Endpoint=sb://some-bus.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=****1Fw=;EntityPath=my-test-queue",
			"project_id": "",
			"queue_url": "",
			"region": "",
			"topic": "",
			"topic_arn": "",
			"type": "AzureServiceBus",
			"uri": ""
		  }
		],
		"format": [
		  {
			"cloud_events_version": "",
			"type": "Platform"
		  }
		],
		"id": "447b287d-e196-433c-b8ef-b858511b61ff",
		"key": "my-subscription-key",
		"message": [
		  {
			"resource_type_id": "product",
			"types": [
			  "ProductCreated"
			]
		  }
		],
		"version": 4
	  }
	`)

	expected := Subscription{
		Version: types.Int64Value(4),
		ID:      types.StringValue("447b287d-e196-433c-b8ef-b858511b61ff"),
		Key:     types.StringValue("my-subscription-key"),
		Changes: []Changes{
			{
				ResourceTypeIds: []types.String{
					types.StringValue("product"),
				},
			},
		},
		Destination: &Destination{
			Type:             types.StringValue("AzureServiceBus"),
			TopicARN:         types.StringValue(""),
			AccessKey:        types.StringValue(""),
			AccessSecret:     types.StringValue(""),
			QueueURL:         types.StringValue(""),
			AccountID:        types.StringValue(""),
			Region:           types.StringValue(""),
			URI:              types.StringValue(""),
			ConnectionString: types.StringValue("Endpoint=sb://some-bus.servicebus.windows.net/;SharedAccessKeyName=test;SharedAccessKey=****1Fw=;EntityPath=my-test-queue"),
			ProjectID:        types.StringValue(""),
			Topic:            types.StringValue(""),
		},
		Format: &Format{
			Type:              types.StringValue("Platform"),
			CloudEventVersion: types.StringValue(""),
		},
		Messages: []Message{
			{
				ResourceTypeID: types.StringValue("product"),
				Types: []types.String{
					types.StringValue("ProductCreated"),
				},
			},
		},
	}

	ctx := context.Background()
	req := resource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{
			JSON: oldState,
		},
	}
	resp := resource.UpgradeStateResponse{}
	upgradeStateV1(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics.Errors())
	require.NotNil(t, resp.DynamicValue)

	// Create the state based on the current schema
	s := getCurrentSchema()
	upgradedStateValue, err := resp.DynamicValue.Unmarshal(s.Type().TerraformType(ctx))
	require.NoError(t, err)
	state := tfsdk.State{
		Raw:    upgradedStateValue,
		Schema: s,
	}

	res := Subscription{}
	diags := state.Get(ctx, &res)
	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, expected, res)
}

func getCurrentSchema() schema.Schema {
	ctx := context.Background()
	res := NewSubscriptionResource()

	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	res.Schema(ctx, req, &resp)
	return resp.Schema
}
