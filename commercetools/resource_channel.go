package commercetools

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/channels"
)

func resourceChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceChannelCreate,
		Read:   resourceChannelRead,
		Update: resourceChannelUpdate,
		Delete: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceChannelCreate(d *schema.ResourceData, m interface{}) error {
	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	roles := []channels.ChannelRole{}
	for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
		roles = append(roles, channels.ChannelRole(value))
	}

	draft := &channels.ChannelDraft{
		Key:         d.Get("key").(string),
		Roles:       roles,
		Name:        name,
		Description: description,
	}

	svc := getChannelService(m)
	channel, err := svc.Create(draft)
	if err != nil {
		return err
	}

	d.SetId(channel.ID)
	d.Set("version", channel.Version)
	return resourceChannelRead(d, m)
}

func resourceChannelRead(d *schema.ResourceData, m interface{}) error {
	svc := getChannelService(m)

	channel, err := svc.GetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			if ctErr.Code() == commercetools.ErrResourceNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(channel.ID)
	d.Set("version", channel.Version)
	d.Set("name", channel.Name)
	d.Set("description", channel.Description)
	d.Set("roles", channel.Roles)
	return nil
}

func resourceChannelUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getChannelService(m)

	input := &channels.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&channels.ChangeName{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&channels.ChangeDescription{Description: newDescription})
	}

	if d.HasChange("roles") {
		roles := []channels.ChannelRole{}
		for _, value := range expandStringArray(d.Get("roles").([]interface{})) {
			roles = append(roles, channels.ChannelRole(value))
		}
		input.Actions = append(
			input.Actions,
			&channels.SetRoles{Roles: roles})
	}

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceChannelRead(d, m)
}

func resourceChannelDelete(d *schema.ResourceData, m interface{}) error {
	svc := getChannelService(m)
	version := d.Get("version").(int)
	_, err := svc.Delete(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func getChannelService(m interface{}) *channels.Service {
	client := m.(*commercetools.Client)
	svc := channels.New(client)
	return svc
}
