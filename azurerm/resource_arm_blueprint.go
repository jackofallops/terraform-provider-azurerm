package azurerm

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"net/http"
	"strings"
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
				ValidateFunc: azure.ValidateResourceID,
				// todo Scope validation function
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
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
						"target_scope": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(blueprint.Subscription),
								string(blueprint.ManagementGroup),
							}, false),
						},
						"versions": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
					},
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceArmBlueprintCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).blueprint
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	bpType := d.Get("type").(string)
	propsRaw := d.Get("properties").([]interface{})
	properties := expandBlueprintProperties(propsRaw)

	scope := d.Get("scope").(string)
	model := blueprint.Model{
		Properties: properties,
		Name:       utils.String(name),
		Type:       utils.String(bpType),
	}

	log.Printf("[SJDEBUG] Model is: %#v", model)
	log.Printf("[SJDEBUG] properties are: %#v", model.Properties)
	log.Printf("[SJDEBUG] Scope is %#v", scope)
	log.Printf("[SJDEBUG] name is %#v", *model.Name)
	log.Printf("[SJDEBUG] type is %#v", *model.Type)

	read, err := client.CreateOrUpdate(ctx, scope, name, model)

	if err != nil {
		return fmt.Errorf("Error creating or updating blueprint: %+v", err)
	}

	d.SetId(*read.ID)
	return resourceArmBlueprintRead(d, meta)
}

func resourceArmBlueprintRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).blueprint
	ctx := meta.(*ArmClient).StopContext

	id, err := azure.ParseAzureResourceID(d.Id())

	if err != nil {
		return err
	}

	scope := parseBlueprintScope(d.Id())
	name := id.Path["blueprints"]

	resp, err := client.Get(ctx, scope, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] blueprint %q was not found in scope %q", name, scope)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading blueprint %q in scope %q", name, scope)
	}

	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("display_name", resp.DisplayName)
	d.Set("description", resp.Description)
	d.Set("properties.parameters", resp.Properties.Parameters)

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
		return fmt.Errorf("Error deleting Blueprint %s in scope %s", name, scope)
	}
	return nil
}

func parseBlueprintScope(input string) (scope string) {
	if strings.HasPrefix(input, "/subscriptions") {
		scope = input[0:52]
	}
	if strings.HasPrefix(input, "/providers/Microsoft.Management/managementGroups/") {
		scope = input[0:86]
	}

	return scope
}

func expandBlueprintProperties(input []interface{}) *blueprint.Properties {
	if len(input) == 0 {
		emptyProps := blueprint.Properties{}
		return &emptyProps
	}
	p := input[0].(map[string]interface{})
	log.Printf("[SJDEBUG] p is: %#v", p)
	ret := blueprint.Properties{}

	if displayName, ok := p["display_name"]; ok {
		log.Printf("[SJDEBUG] Display name: %#v", displayName)
		ret.DisplayName = utils.String(displayName.(string))
	}

	if description, ok := p["description"]; ok {
		log.Printf("[SJDEBUG] Description: %#v", description)
		ret.Description = utils.String(description.(string))
	}
	if layout, ok := p["layout"]; ok {
		ret.Layout = layout
	} else {
		ret.Layout = ""
	}
	if statRaw, ok := p["status"]; ok {
		status := statRaw.(*blueprint.Status)
		ret.Status = status
	}

	if ts, ok := p["target_scope"]; ok {
		switch ts {
		case "subscription":
			ret.TargetScope = blueprint.Subscription
		case "managementGroup":
			ret.TargetScope = blueprint.ManagementGroup
		}
		log.Printf("[SJDEBUG] targetScope is %#v", ret.TargetScope)
	}
	pdm := map[string]*blueprint.ParameterDefinition{}

	if params, ok := p["parameters"].(map[string]*blueprint.ParameterDefinition); ok {
		for k, v := range params {
			pdm[k] = v
			log.Printf("[SJDEBUG] params parsed")
		}
	} //else {
	//	ret.Parameters = nil
	//}
	if rg, ok := p["resource_groups"]; ok {
		//todo handle resource groups in  props expansion
		log.Printf("[SJDEBUG] RG is: %#v", rg)
		ret.ResourceGroups = map[string]*blueprint.ResourceGroupDefinition{}
	} else {
		ret.ResourceGroups = map[string]*blueprint.ResourceGroupDefinition{}
	}
	if ver, ok := p["versions"]; ok {
		//todo handle versions in props expansion
		log.Printf("[SJDEBUG] versions are: %#v", ver)
		ret.Versions = []string{}
	} else {
		ret.Versions = []string{}
	}
	log.Printf("Ret value: %#v", ret)
	return &ret
}
