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
				Optional: true,
			},
			"currencies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"version": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	log.Fatal("A project can not be created through terraform")
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
	d.Set("currencies", project.Currencies)
	// d.Set("countries", project.Countries)
	// d.Set("languages", project.Languages)
	// d.Set("createdAt", project.CreatedAt)
	// d.Set("trialUntil", project.TrialUntil)
	// d.Set("messages", project.Messages)
	// d.Set("shippingRateInputType", project.ShippingRateInputType)

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getProjectService(m)

	input := &project.UpdateInput{
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("currencies") {
		newCurrencies := resourceProjectGetCurrencies(d)
		input.Actions = append(
			input.Actions,
			&project.ChangeCurrencies{Currencies: newCurrencies})
	}

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	log.Fatal("A project can not be deleted through terraform")
	return fmt.Errorf("A project can not be deleted through terraform")
}

func getProjectService(m interface{}) *project.Service {
	client := m.(*commercetools.Client)
	svc := project.New(client)
	return svc
}

func resourceProjectGetCurrencies(d *schema.ResourceData) []string {
	input := d.Get("currencies").([]interface{})
	var currencyObjects []string
	for _, raw := range input {
		currencyObjects = append(currencyObjects, raw.(string))
	}

	return currencyObjects
}
