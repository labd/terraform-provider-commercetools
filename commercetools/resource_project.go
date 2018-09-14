package commercetools

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/project"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"currencies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"countries": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"languages": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"messages": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"version": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	log.Print("A project can not be created through terraform")
	return fmt.Errorf("A project can not be created through terraform")
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading projects from commercetools")
	svc := getProjectService(m)

	project, err := svc.Get()

	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			if ctErr.Code() == commercetools.ErrResourceNotFound {
				return nil
			}
		}
		return err
	}

	log.Print("[DEBUG] Found the following project:")
	log.Print(stringFormatObject(project))

	d.SetId(project.Key)
	d.Set("version", project.Version)
	d.Set("name", project.Name)
	d.Set("currencies", project.Currencies)
	d.Set("countries", project.Countries)
	d.Set("languages", project.Languages)
	// d.Set("createdAt", project.CreatedAt)
	// d.Set("trialUntil", project.TrialUntil)
	log.Print("[DEBUG] Logging messages enabled")
	log.Print(stringFormatObject(project.Messages))
	d.Set("messages", project.Messages)
	log.Print(stringFormatObject(d))
	// d.Set("shippingRateInputType", project.ShippingRateInputType)

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getProjectService(m)

	input := &project.UpdateInput{
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("name") {
		input.Actions = append(input.Actions, &project.ChangeName{d.Get("name").(string)})
	}

	if d.HasChange("currencies") {
		newCurrencies := getStringSlice(d, "currencies")
		input.Actions = append(
			input.Actions,
			&project.ChangeCurrencies{Currencies: newCurrencies})
	}

	if d.HasChange("countries") {
		newCountries := getStringSlice(d, "countries")
		input.Actions = append(
			input.Actions,
			&project.ChangeCountries{Countries: newCountries})
	}

	if d.HasChange("languages") {
		newLanguages := getStringSlice(d, "languages")
		input.Actions = append(
			input.Actions,
			&project.ChangeLanguages{Languages: newLanguages})
	}

	if d.HasChange("messages") {
		messages := d.Get("messages").(map[string]interface{})
		// ¯\_(ツ)_/¯
		enabled := false
		if messages["enabled"] == "1" {
			enabled = true
		}

		input.Actions = append(
			input.Actions,
			&project.ChangeMessagesEnabled{MessagesEnabled: enabled})
	}

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	log.Print("A project can not be deleted through terraform")
	return fmt.Errorf("A project can not be deleted through terraform")
}

func getProjectService(m interface{}) *project.Service {
	client := m.(*commercetools.Client)
	svc := project.New(client)
	return svc
}

func getStringSlice(d *schema.ResourceData, field string) []string {
	input := d.Get(field).([]interface{})
	var currencyObjects []string
	for _, raw := range input {
		currencyObjects = append(currencyObjects, raw.(string))
	}

	return currencyObjects
}
