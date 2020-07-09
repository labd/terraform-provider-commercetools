package commercetools

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceStateTransitions() *schema.Resource {
	return &schema.Resource{
		Create: resourceStateTransitionsCreate,
		Read:   resourceStateTransitionsRead,
		Update: resourceStateTransitionsUpdate,
		Delete: resourceStateTransitionsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"from": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"to": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceStateTransitionsCreate(d *schema.ResourceData, m interface{}) error {
	fromStateID := d.Get("from").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(fromStateID)
	defer ctMutexKV.Unlock(fromStateID)

	client := getClient(m)
	log.Printf("[DEBUG] Reading state from commercetools, with id: %s", fromStateID)
	fromState, err := client.StateGetWithID(fromStateID)
	if err != nil {
		return err
	}

	updatedState, err := updateTransitions(client, stateTransitionsUpdateInput{
		ID:            fromState.ID,
		Version:       fromState.Version,
		TransitionIDs: d.Get("to").(*schema.Set).List(),
	})
	if err != nil {
		return err
	}

	d.SetId(updatedState.ID)
	return setStateTransitionsResourceState(d, updatedState)
}

func resourceStateTransitionsRead(d *schema.ResourceData, m interface{}) error {
	fromStateID := d.Id()
	client := getClient(m)
	log.Printf("[DEBUG] Reading state from commercetools, with id: %s", fromStateID)
	fromState, err := client.StateGetWithID(fromStateID)

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] No state found with id: %s", fromStateID)
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if fromState == nil {
		log.Printf("[DEBUG] No state found with id: %s", fromStateID)
		d.SetId("")
	} else {
		log.Printf("[DEBUG] No state found with id: %s", fromStateID)
		log.Print(stringFormatObject(fromState))

		err = setStateTransitionsResourceState(d, fromState)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceStateTransitionsUpdate(d *schema.ResourceData, m interface{}) error {
	fromStateID := d.Get("from").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(fromStateID)
	defer ctMutexKV.Unlock(fromStateID)

	client := getClient(m)
	log.Printf("[DEBUG] Reading state from commercetools, with id: %s", fromStateID)
	fromState, err := client.StateGetWithID(fromStateID)

	if err != nil {
		return err
	}

	if d.HasChange("to") {
		updatedState, err := updateTransitions(client, stateTransitionsUpdateInput{
			ID:            fromStateID,
			Version:       fromState.Version,
			TransitionIDs: d.Get("to").(*schema.Set).List(),
		})
		if err != nil {
			return err
		}

		return setStateTransitionsResourceState(d, updatedState)
	}

	return nil
}

func resourceStateTransitionsDelete(d *schema.ResourceData, m interface{}) error {
	fromStateID := d.Get("from").(string)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(fromStateID)
	defer ctMutexKV.Unlock(fromStateID)

	client := getClient(m)
	log.Printf("[DEBUG] Reading state from commercetools, with id: %s", fromStateID)
	fromState, err := client.StateGetWithID(fromStateID)

	if err != nil {
		return err
	}

	// TODO: How do we indicate that we should be destroyed before other
	// transitions are created for this state?
	_, err = updateTransitions(client, stateTransitionsUpdateInput{
		ID:            fromState.ID,
		Version:       fromState.Version,
		TransitionIDs: nil,
	})
	if ctErr, ok := err.(commercetools.ErrorResponse); ok {
		if ctErr.StatusCode == 404 {
			return nil // Ignore a 404, the state is destroyed so we have been destroyed implicitly
		}
	}
	return err
}

func setStateTransitionsResourceState(d *schema.ResourceData, state *commercetools.State) error {
	if state.Transitions == nil {
		d.SetId("")
		return nil
	}

	toStates := make([]string, len(state.Transitions))
	for i, stateRef := range state.Transitions {
		toStates[i] = stateRef.ID
	}

	d.Set("from", state.ID)
	d.Set("to", toStates)

	log.Printf("[DEBUG] New state: %#v", d)
	return nil
}

// Commerce Tools client patch
//
// There is a BIG difference between states with transitions = null vs transitions = []
// A state with transitions = null is one that can be transitioned into any other state
// A state with transitions = [] is one that can not be transitioned at all
//
// Unfortunately, the Commerce Tools go SDK omits empty transitions from the setTransitions
// update action, meaning that an empty array of transitions or a null transitions array
// will both result in null transitions in Commerce Tools.
//
// Would need to tackle this https://github.com/labd/commercetools-go-sdk/pull/45 if we
// wanted to fix this upstream.

type stateTransitionsUpdateInput struct {
	ID            string
	Version       int
	TransitionIDs []interface{}
}

// Our patched version of commercetools.StateSetTransitionsAction
type stateSetTransitionsAction struct {
	Transitions []commercetools.StateResourceIdentifier `json:"transitions"`
}

// MarshalJSON override to set the discriminator value
func (obj stateSetTransitionsAction) MarshalJSON() ([]byte, error) {
	type Alias stateSetTransitionsAction
	return json.Marshal(struct {
		Action string `json:"action"`
		*Alias
	}{Action: "setTransitions", Alias: (*Alias)(&obj)})
}

func updateTransitions(client *commercetools.Client, input stateTransitionsUpdateInput) (result *commercetools.State, err error) {
	var transitions []commercetools.StateResourceIdentifier
	if input.TransitionIDs != nil {
		transitions = make([]commercetools.StateResourceIdentifier, len(input.TransitionIDs))
		for i, value := range input.TransitionIDs {
			transitions[i] = commercetools.StateResourceIdentifier{
				ID: value.(string),
			}
		}
	}

	actions := []interface{}{stateSetTransitionsAction{Transitions: transitions}}
	endpoint := strings.Replace("states/{ID}", "{ID}", input.ID, 1)

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		attemptErr := client.Update(endpoint, nil, input.Version, actions, &result)
		if attemptErr != nil {
			return handleCommercetoolsError(attemptErr)
		}
		return nil
	})

	return
}
