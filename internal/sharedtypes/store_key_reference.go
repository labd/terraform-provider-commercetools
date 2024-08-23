package sharedtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
)

var (
	StoreKeyReferenceBlockSchema = schema.ListNestedBlock{
		MarkdownDescription: "Sets the Stores the Business Unit is associated with. \n\nIf the Business Unit has Stores defined, " +
			"then all of its Carts, Orders, Quotes, or Quote Requests must belong to one of the Business Unit's " +
			"Stores.\n\nIf the Business Unit has no Stores, then all of its Carts, Orders, Quotes, or Quote Requests " +
			"must not belong to any Store.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{
					MarkdownDescription: "User-defined unique identifier of the Store",
					Optional:            true,
				},
			},
		},
	}
)

// StoreKeyReference is a type to model the fields that all types of StoreKeyReference have in common.
type StoreKeyReference struct {
	Key types.String `tfsdk:"key"`
}

func NewStoreKeyReferenceFromNative(n *platform.StoreKeyReference) StoreKeyReference {
	return StoreKeyReference{
		Key: types.StringValue(n.Key),
	}
}
