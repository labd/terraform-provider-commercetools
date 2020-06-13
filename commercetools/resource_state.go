package commercetools

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceState() *schema.Resource {
	return &schema.Resource{
		Create: resourceStateCreate,
		Read:   resourceStateRead,
		Update: resourceStateUpdate,
		Delete: resourceStateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			// TODO: Now that we've added resource_state_transitions we can not guarantee
			// this version will stay in sync. Should we remove it altogether?
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(commercetools.StateTypeEnumOrderState),
					string(commercetools.StateTypeEnumLineItemState),
					string(commercetools.StateTypeEnumProductState),
					string(commercetools.StateTypeEnumReviewState),
					string(commercetools.StateTypeEnumPaymentState),
				}, false),
			},
			"name": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"initial": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						string(commercetools.StateRoleEnumReviewIncludedInStatistics),
						string(commercetools.StateRoleEnumReturn),
					}, false),
				},
			},
			"transitions": {
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

		state, err = client.StateCreate(draft)
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
	stateID := d.Id()
	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(stateID)
	defer ctMutexKV.Unlock(stateID)

	client := getClient(m)
	state, err := client.StateGetWithID(stateID)

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
	stateID := d.Id()
	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(stateID)
	defer ctMutexKV.Unlock(stateID)

	client := getClient(m)

	input := &commercetools.StateUpdateWithIDInput{
		ID:      stateID,
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

	_, err := client.StateUpdateWithID(input)
	if err != nil {
		return err
	}

	return resourceStateRead(d, m)
}

func resourceStateDelete(d *schema.ResourceData, m interface{}) error {
	stateID := d.Id()
	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(stateID)
	defer ctMutexKV.Unlock(stateID)

	client := getClient(m)
	state, err := client.StateGetWithID(stateID)
	// TODO: Do we need to handle a 404 in either of these methods?

	if err != nil {
		return err
	}

	_, err = client.StateDeleteWithID(stateID, state.Version)
	return err
}
