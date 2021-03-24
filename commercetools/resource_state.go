package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceState() *schema.Resource {
	return &schema.Resource{
		Description: "The commercetools platform allows you to model states of certain objects, such as orders, line " +
			"items, products, reviews, and payments to define finite state machines reflecting the business " +
			"logic you'd like to implement.\n\n" +
			"See also the [State API Documentation](https://docs.commercetools.com/api/projects/states)",
		Create: resourceStateCreate,
		Read:   resourceStateRead,
		Update: resourceStateUpdate,
		Delete: resourceStateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "A unique identifier for the state",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Description: "[StateType](https://docs.commercetools.com/api/projects/states#statetype)",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					string(commercetools.StateTypeEnumOrderState),
					string(commercetools.StateTypeEnumLineItemState),
					string(commercetools.StateTypeEnumProductState),
					string(commercetools.StateTypeEnumReviewState),
					string(commercetools.StateTypeEnumPaymentState),
				}, false),
			},
			"name": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"description": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
			},
			"initial": {
				Description: "A state can be declared as an initial state for any state machine. When a workflow " +
					"starts, this first state must be an initial state",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"roles": {
				Description: "Array of [State Role](https://docs.commercetools.com/api/projects/states#staterole)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(commercetools.StateRoleEnumReviewIncludedInStatistics),
						string(commercetools.StateRoleEnumReturn),
					}, false),
				},
			},
			"transitions": {
				Description: "Transitions are a way to describe possible transformations of the current state to other " +
					"states of the same type (for example: Initial -> Shipped). When performing a transitionState update " +
					"action and transitions is set, the currently referenced state must have a transition to the new state.\n" +
					"If transitions is an empty list, it means the current state is a final state and no further " +
					"transitions are allowed.\nIf transitions is not set, the validation is turned off. When " +
					"performing a transitionState update action, any other state of the same type can be transitioned to",
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceStateCreate(d *schema.ResourceData, m interface{}) error {
	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	roles := []commercetools.StateRoleEnum{}
	for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
		roles = append(roles, commercetools.StateRoleEnum(value))
	}

	var transitions []commercetools.StateResourceIdentifier
	for _, value := range d.Get("transitions").(*schema.Set).List() {
		transitions = append(transitions, commercetools.StateResourceIdentifier{
			Key: value.(string),
		})
	}

	draft := &commercetools.StateDraft{
		Key:         d.Get("key").(string),
		Type:        commercetools.StateTypeEnum(d.Get("type").(string)),
		Name:        &name,
		Description: &description,
		Roles:       roles,
		Transitions: transitions,
	}

	// Note the use of GetOkExists since it's an optional bool type
	if _, exists := d.GetOkExists("initial"); exists {
		draft.Initial = d.Get("initial").(bool)
	}

	client := getClient(m)
	var state *commercetools.State

	err := resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error

		state, err = client.StateCreate(context.Background(), draft)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.SetId(state.ID)
	d.Set("version", state.Version)
	return resourceStateRead(d, m)
}

func resourceStateRead(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	state, err := client.StateGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(state.ID)
	d.Set("version", state.Version)
	d.Set("key", state.Key)
	d.Set("type", state.Type)
	if state.Name != nil {
		d.Set("name", *state.Name)
	}
	if state.Description != nil {
		d.Set("description", *state.Description)
	}
	d.Set("initial", state.Initial)
	if state.Roles != nil {
		d.Set("roles", state.Roles)
	}
	if state.Transitions != nil {
		d.Set("transitions", state.Transitions)
	}
	return nil
}

func resourceStateUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.StateUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.StateUpdateAction{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.StateSetNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescription := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.StateSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.StateChangeKeyAction{Key: newKey})
	}

	if d.HasChange("type") {
		newType := d.Get("type").(commercetools.StateTypeEnum)
		input.Actions = append(
			input.Actions,
			&commercetools.StateChangeTypeAction{Type: newType})
	}

	if d.HasChange("initial") {
		newInitial := d.Get("initial").(bool)
		input.Actions = append(
			input.Actions,
			&commercetools.StateChangeInitialAction{Initial: newInitial})
	}

	if d.HasChange("roles") {
		roles := []commercetools.StateRoleEnum{}
		for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
			roles = append(roles, commercetools.StateRoleEnum(value))
		}
		input.Actions = append(
			input.Actions,
			&commercetools.StateSetRolesAction{Roles: roles})
	}

	if d.HasChange("transitions") {
		var transitions []commercetools.StateResourceIdentifier
		for _, value := range d.Get("transitions").(*schema.Set).List() {
			transitions = append(transitions, commercetools.StateResourceIdentifier{
				Key: value.(string),
			})
		}
		input.Actions = append(
			input.Actions,
			&commercetools.StateSetTransitionsAction{
				Transitions: transitions,
			})
	}

	_, err := client.StateUpdateWithID(context.Background(), input)
	if err != nil {
		return err
	}

	return resourceStateRead(d, m)
}

func resourceStateDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.StateDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}
