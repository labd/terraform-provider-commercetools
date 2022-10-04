package commercetools

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

var stateTransitionIds map[string]bool

func resourceStateTransitions() *schema.Resource {
	return &schema.Resource{
		Description: "Transitions are a way to describe possible transformations of the current state to other " +
			"states of the same type (for example: Initial -> Shipped). When performing a transitionState update " +
			"action and transitions is set, the currently referenced state must have a transition to the new state.\n" +
			"If transitions is an empty list, it means the current state is a final state and no further " +
			"transitions are allowed.\nIf transitions is not set, the validation is turned off. When " +
			"performing a transitionState update action, any other state of the same type can be transitioned to.\n\n" +
			"Note: Only one resource can be created for each state",
		CreateContext: resourceStateTransitionsCreate,
		ReadContext:   resourceStateTransitionsRead,
		UpdateContext: resourceStateTransitionsUpdate,
		DeleteContext: resourceStateTransitionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceStateTransitionsImportState,
		},
		Schema: map[string]*schema.Schema{
			"from": {
				Description: "ID of the state to transition from",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: func(val any, key string) ([]string, []error) {
					ID := val.(string)

					ctMutexKV.Lock("stateTransitionsIds")
					defer ctMutexKV.Unlock("stateTransitionsIds")

					if stateTransitionIds == nil {
						stateTransitionIds = make(map[string]bool)
					}

					if _, exists := stateTransitionIds[ID]; exists {
						return []string{
							fmt.Sprintf("can only define one state transitions resource for state with ID %s", ID),
						}, nil
					}
					stateTransitionIds[ID] = true
					return nil, nil
				},
			},
			"to": {
				Description: "Transitions are a way to describe possible transformations of the current state to other " +
					"states of the same type (for example: Initial -> Shipped). When performing a transitionState update " +
					"action and transitions is set, the currently referenced state must have a transition to the new state.\n" +
					"If transitions is an empty list, it means the current state is a final state and no further " +
					"transitions are allowed.\nIf transitions is not set, the validation is turned off. When " +
					"performing a transitionState update action, any other state of the same type can be transitioned to",
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceStateTransitionsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	ID := d.Get("from").(string)

	transitions := expandStateTransitions(d)
	state, err := resourceStateSetTransitions(ctx, client, ID, transitions)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	d.SetId(state.ID)
	d.Set("to", flattenStateTransitions(state.Transitions))
	return nil
}

func resourceStateTransitionsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	ID := d.Get("from").(string)

	state, err := client.States().WithId(ID).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(state.ID)
	d.Set("to", flattenStateTransitions(state.Transitions))
	return nil
}

func resourceStateTransitionsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	ID := d.Get("from").(string)

	if d.HasChange("to") {
		transitions := expandStateTransitions(d)
		state, err := resourceStateSetTransitions(ctx, client, ID, transitions)
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		d.Set("to", flattenStateTransitions(state.Transitions))
	}

	return nil
}

func resourceStateTransitionsDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	ID := d.Get("from").(string)

	var transitions []platform.StateResourceIdentifier
	_, err := resourceStateSetTransitions(ctx, client, ID, transitions)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}
	return nil
}

func resourceStateTransitionsImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := getClient(meta)
	state, err := client.States().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		return nil, err
	}

	data := resourceStateTransitions().Data(nil)
	data.SetId(state.ID)
	data.Set("from", state.ID)
	data.Set("to", flattenStateTransitions(state.Transitions))

	result := []*schema.ResourceData{data}
	return result, nil
}

func resourceStateSetTransitions(ctx context.Context, client *platform.ByProjectKeyRequestBuilder, stateId string, transitions []platform.StateResourceIdentifier) (*platform.State, error) {
	state, err := client.States().WithId(stateId).Get().Execute(ctx)
	if err != nil {
		return nil, err
	}

	input := platform.StateUpdate{
		Version: state.Version,
		Actions: []platform.StateUpdateAction{},
	}

	// Validate that the transitions are modified before trying to update them.
	// If we try to set the transitions to a value already set then commercetools
	// returns an InvalidOperation with `'transitions' has no changes.`
	// Normally this should never happen, but since we moved to a separate
	// resource in 1.5.0 this does occur because we create a new resource to
	// set the value set before on the state resource.
	// See issue #312
	newTransitionIds := make([]string, len(transitions))
	for i, t := range transitions {
		newTransitionIds[i] = *t.ID
	}

	curTransitionIds := make([]string, len(state.Transitions))
	for i, t := range state.Transitions {
		curTransitionIds[i] = t.ID
	}

	if reflect.DeepEqual(newTransitionIds, curTransitionIds) {
		log.Println("Transitions not modified, ignoring updating")
		return state, nil
	}

	input.Actions = append(
		input.Actions,
		&platform.StateSetTransitionsAction{
			Transitions: transitions,
		})

	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.States().WithId(stateId).Post(input).Execute(ctx)
		return processRemoteError(err)
	})

	return state, err
}

func expandStateTransitions(d *schema.ResourceData) []platform.StateResourceIdentifier {
	values := d.Get("to").(*schema.Set).List()
	transitions := make([]platform.StateResourceIdentifier, len(values))
	for i, value := range values {
		transitions[i] = platform.StateResourceIdentifier{
			ID: stringRef(value),
		}
	}
	return transitions
}

func flattenStateTransitions(values []platform.StateReference) []string {
	result := make([]string, len(values))
	for i := range values {
		result[i] = values[i].ID
	}
	return result
}
