package azurerm

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Type:     schema.TypeString,
				Required: true,
				// todo Scope validation function to cover managementGroup condition?
				ValidateFunc: validateBlueprintScope,
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
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(blueprint.Subscription),
								string(blueprint.ManagementGroup),
							}, false),
						},
						"versions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"status": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"last_modified": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"time_created": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
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
	client := meta.(*ArmClient).Blueprint.BlueprintsClient
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
		ID:         utils.String(""),
	}

	read, err := client.CreateOrUpdate(ctx, scope, name, model)

	if err != nil {
		return fmt.Errorf("Error creating or updating blueprint: %+v", err)
	}

	d.SetId(*read.ID)
	return resourceArmBlueprintRead(d, meta)
}

func resourceArmBlueprintRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Blueprint.BlueprintsClient
	ctx := meta.(*ArmClient).StopContext

	// Can't use ParseAzureResourceID as normal, as management group id value doesn't start "/subscriptions"
	scope := d.Get("scope").(string)
	name := d.Get("name").(string)

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
	d.Set("scope", scope)
	props := flattenBlueprintProperties(resp.Properties)
	d.Set("properties", props)

	return nil
}

func resourceArmBlueprintDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Blueprint.BlueprintsClient
	ctx := meta.(*ArmClient).StopContext

	scope := d.Get("scope").(string)
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

func expandBlueprintProperties(input []interface{}) *blueprint.Properties {
	if len(input) == 0 {
		emptyProps := blueprint.Properties{}
		return &emptyProps
	}

	p := input[0].(map[string]interface{})

	ret := blueprint.Properties{}

	if displayName, ok := p["display_name"]; ok {
		ret.DisplayName = utils.String(displayName.(string))
	}

	if description, ok := p["description"]; ok {
		ret.Description = utils.String(description.(string))
	}

	if layout, ok := p["layout"]; ok {
		ret.Layout = layout
	} else {
		ret.Layout = map[string]string{}
	}

	if ts, ok := p["target_scope"]; ok {
		switch ts {
		case "subscription":
			ret.TargetScope = blueprint.Subscription
		case "managementGroup":
			ret.TargetScope = blueprint.ManagementGroup
		}
	}

	pdm := map[string]*blueprint.ParameterDefinition{}

	if params, ok := p["parameters"].(map[string]*blueprint.ParameterDefinition); ok {
		for k, v := range params {
			pdm[k] = v
		}
	} else {
		ret.Parameters = map[string]*blueprint.ParameterDefinition{}
	}

	if _, ok := p["resource_groups"]; ok {
		//todo handle resource groups in  props expansion
		ret.ResourceGroups = map[string]*blueprint.ResourceGroupDefinition{}
	} else {
		ret.ResourceGroups = map[string]*blueprint.ResourceGroupDefinition{}
	}

	if _, ok := p["versions"]; ok {
		// todo - handle Versions object when I figure out structure
		ret.Versions = map[string]string{}
	} else {
		ret.Versions = map[string]string{}
	}

	return &ret
}

func flattenBlueprintProperties(blueprintProperties *blueprint.Properties) []interface{} {

	props := make(map[string]interface{})

	props["display_name"] = &blueprintProperties.DisplayName
	props["description"] = &blueprintProperties.Description
	props["target_scope"] = string(blueprintProperties.TargetScope)
	props["resource_group"] = &blueprintProperties.ResourceGroups
	props["parameters"] = &blueprintProperties.Parameters

	return []interface{}{props}
}

func validateBlueprintScope(i interface{}, k string) (warnings []string, errors []error) {
	input := i.(string)

	if strings.HasPrefix(input, "/subscription") {
		_, err := azure.ValidateResourceID(i, input)
		if err != nil {
			errors = append(errors, fmt.Errorf("Subscription specified is not a valid Resource ID: %q", k))
		}
	} else if strings.HasPrefix(input, "/providers/Microsoft.Management/managementGroups/") {
		input = strings.TrimPrefix(input, "/")
		input = strings.TrimSuffix(input, "/")
		components := strings.Split(input, "/")

		if len(components) != 4 {
			errors = append(errors, fmt.Errorf("Invalid management group path, should contain 4 elements not %q", len(components)))
		}
		_, err := validate.UUID(components[3], input)
		if err != nil {
			errors = append(errors, fmt.Errorf("Management group ID not a valid uuid: %q", components[3]))
		}
	} else {
		errors = append(errors, fmt.Errorf("Invalid scope, should be a subscription resource ID or Management Group path: %q", k))
	}

	return warnings, errors
}
