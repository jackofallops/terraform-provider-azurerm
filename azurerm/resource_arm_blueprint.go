package azurerm

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"net/http"
)

func resourceArmBlueprint() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmBlueprintCreateOrUpdate,
		Read:   resourceArmBlueprintRead,
		Update: resourceArmBlueprintCreateOrUpdate,
		Delete: resourceArmBlueprintDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:         schema.TypeString,
				Required:     true,
				// todo Scope validation function
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"properties": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameters": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.ValidateJsonString,
						},
						"resource_groups": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.ValidateJsonString,
						},
						"targetScope": {
							Type:     schema.TypeString,
							Required: true,
							Default:  blueprint.Subscription,
							ValidateFunc: validation.StringInSlice([]string{
								string(blueprint.Subscription),
								string(blueprint.ManagementGroup),
							}, true),
						},
						"versions": {
							Type:     schema.TypeSet,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceArmBlueprintCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).blueprint
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	dName := d.Get("display_name").(string)
	desc := d.Get("description").(string)

	bp := blueprint.Properties{
		DisplayName: &dName,
		Description: &desc,
		TargetScope: d.Get("target_scope").(blueprint.TargetScope),
	}

	scope := d.Get("subscription").(string)

	model := blueprint.Model{
		Properties: &bp,
	}

	model, err := client.CreateOrUpdate(ctx, scope, name, model)
	if err != nil {

	}
	return resourceArmBlueprintRead(d, meta)
}

func resourceArmBlueprintRead(d *schema.ResourceData, meta interface{}) error {

	return nil
}
func resourceArmBlueprintDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).blueprint
	ctx := meta.(*ArmClient).StopContext

	scope := d.Get("subscription").(string)
	name := d.Get("name").(string)
	resp, err := client.Delete(ctx, scope, name)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error deleting Blueprint %s in subscription %s", name, scope)
	}
	return nil
}
