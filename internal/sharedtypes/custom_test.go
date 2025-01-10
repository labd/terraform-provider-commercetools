package sharedtypes

import (
	"fmt"
	"testing"

	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

var nilEncoder CustomFieldTypeEncoder = func(t *platform.Type, name string, value any) (any, error) {
	return nil, nil
}

func TestCustomIsSet(t *testing.T) {
	c := &Custom{TypeID: nil}
	assert.False(t, c.IsSet())

	c = &Custom{TypeID: new(string)}
	assert.True(t, c.IsSet())
}

func TestCustomFieldsInterface(t *testing.T) {
	c := &Custom{Fields: map[string]string{"key1": "value1", "key2": "value2"}}
	expected := map[string]any{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expected, c.fieldsInterface())
}

func TestCustomDraftWithEncoder(t *testing.T) {
	t.Run("nil type", func(t *testing.T) {
		c := &Custom{}
		draft, err := c.draftWithEncoder(nil, nilEncoder)
		assert.Nil(t, draft)
		assert.NoError(t, err)
	})

	t.Run("valid type", func(t *testing.T) {
		typeID := "type-id"
		c := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		tp := &platform.Type{FieldDefinitions: []platform.FieldDefinition{{Name: "key", Type: platform.CustomFieldStringType{}}}}
		draft, err := c.draftWithEncoder(tp, nilEncoder)
		assert.NotNil(t, draft)
		assert.NoError(t, err)
	})
}

func TestCustomFieldUpdateActionsWithEncoder(t *testing.T) {
	t.Run("remove custom field", func(t *testing.T) {
		actions, err := customFieldUpdateActionsWithEncoder[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](nil, nilEncoder, &Custom{}, nil)
		assert.NoError(t, err)
		assert.Len(t, actions, 1)
	})

	t.Run("add custom field", func(t *testing.T) {
		typeID := "type-id"
		plan := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		tp := &platform.Type{FieldDefinitions: []platform.FieldDefinition{{Name: "key", Type: platform.CustomFieldStringType{}}}}
		actions, err := customFieldUpdateActionsWithEncoder[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](tp, nilEncoder, nil, plan)
		assert.NoError(t, err)
		assert.Len(t, actions, 1)
	})

	t.Run("with changes", func(t *testing.T) {
		typeID := "type-id"
		current := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		plan := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "new value"}}
		tp := &platform.Type{FieldDefinitions: []platform.FieldDefinition{{Name: "key", Type: platform.CustomFieldStringType{}}}}
		u, err := customFieldUpdateActionsWithEncoder[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](tp, func(t *platform.Type, name string, value any) (any, error) {
			return value, nil
		}, current, plan)
		assert.NoError(t, err)
		assert.Len(t, u, 1)
		assert.Equal(t, "new value", u[0].(platform.ChannelSetCustomFieldAction).Value)
	})

	t.Run("without changes", func(t *testing.T) {
		typeID := "type-id"
		current := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		plan := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		tp := &platform.Type{FieldDefinitions: []platform.FieldDefinition{{Name: "key", Type: platform.CustomFieldStringType{}}}}
		u, err := customFieldUpdateActionsWithEncoder[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](tp, func(t *platform.Type, name string, value any) (any, error) {
			return value, nil
		}, current, plan)
		assert.NoError(t, err)
		assert.Len(t, u, 0)
	})

	t.Run("error encoding", func(t *testing.T) {
		typeID := "type-id"
		plan := &Custom{TypeID: &typeID, Fields: map[string]string{"key": "value"}}
		tp := &platform.Type{FieldDefinitions: []platform.FieldDefinition{{Name: "key", Type: platform.CustomFieldStringType{}}}}
		_, err := customFieldUpdateActionsWithEncoder[platform.ChannelSetCustomTypeAction, platform.ChannelSetCustomFieldAction](tp, func(t *platform.Type, name string, value any) (any, error) {
			return nil, fmt.Errorf("error")
		}, nil, plan)
		assert.Error(t, err)
	})
}

func TestNewCustomFromNative(t *testing.T) {
	t.Run("nil custom fields", func(t *testing.T) {
		c, err := NewCustomFromNative(nil)
		assert.NoError(t, err)
		assert.Nil(t, c)
	})

	t.Run("only type id", func(t *testing.T) {
		cf := &platform.CustomFields{Type: platform.TypeReference{ID: "type-id"}, Fields: map[string]any{}}
		c, err := NewCustomFromNative(cf)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, "type-id", *c.TypeID)
		assert.Nil(t, c.Fields)
	})

	t.Run("valid custom fields", func(t *testing.T) {
		cf := &platform.CustomFields{Type: platform.TypeReference{ID: "type-id"}, Fields: map[string]any{
			"key":  "value",
			"key2": map[string]any{"nested": "value"},
		}}
		c, err := NewCustomFromNative(cf)
		assert.NoError(t, err)
		assert.Equal(t, "type-id", *c.TypeID)
		assert.Equal(t, map[string]string{"key": "value", "key2": "{\"nested\":\"value\"}"}, c.Fields)
	})

	t.Run("failed formatting value", func(t *testing.T) {
		//Create a channel
		ch := make(chan int)

		cf := &platform.CustomFields{Type: platform.TypeReference{ID: "type-id"}, Fields: map[string]any{
			"key": ch,
		}}
		_, err := NewCustomFromNative(cf)
		assert.Error(t, err)
	})
}
