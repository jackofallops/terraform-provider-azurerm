package azurerm

import (
	"bytes"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"net/http"
)

func resourceArmBlueprintPolicyAssignmentArtifact() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmBlueprintPolicyAssignmentArtifactCreateOrUpdate,
		Update: resourceArmBlueprintPolicyAssignmentArtifactCreateOrUpdate,
		Delete: resourceArmBlueprintArtifactDelete,
		Read:   resourceArmBlueprintArtifactRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateBlueprintScope,
			},
			"blueprint_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"depends_on": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"policy_definition_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceArmBlueprintPolicyAssignmentArtifactCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Blueprint.ArtifactsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)
	blueprintName := d.Get("blueprint_name").(string)

	kind := blueprint.KindPolicyAssignment

	artifactProperties := expandPolicyAssignmentArtifactProperties(d)

	artifact := blueprint.PolicyAssignmentArtifact{
		Name: &name,
		Kind: kind,
	}
	artifact.PolicyAssignmentArtifactProperties = artifactProperties

	res, err := client.CreateOrUpdate(ctx, scope, blueprintName, name, artifact)
	if res.Response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error creating or updating blueprint artifact: %q", err)
	}

	return err
}

func resourceArmBlueprintArtifactRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Blueprint.ArtifactsClient
	ctx := meta.(*ArmClient).StopContext

	scope := d.Get("scope").(string)
	artifactName := d.Get("name").(string)
	blueprintName := d.Get("blueprint_name").(string)

	resp, err := client.Get(ctx, scope, blueprintName, artifactName)

	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] artifact %q was not found under %v in scope %q", artifactName, blueprintName, scope)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading artifact %v from blueprint %v in scope %v", artifactName, blueprintName, scope)
	}

	artifactRaw := resp.Value.(blueprint.BasicArtifact)
	artifact, _ := artifactRaw.AsPolicyAssignmentArtifact()
	if artifact.Kind != blueprint.KindPolicyAssignment {
		return fmt.Errorf("Bad: artifact kind expected to be policyAssignment, got %q", artifact.Kind)
	}
	d.Set("name", artifact.Name)
	d.Set("scope", scope)
	d.Set("blueprint_name", blueprintName)

	properties := flattenBlueprintPolicyAssignmentProperties(*artifact.PolicyAssignmentArtifactProperties)
	d.Set("properties", properties)

	return nil
}

func resourceArmBlueprintArtifactDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Blueprint.ArtifactsClient
	ctx := meta.(*ArmClient).StopContext

	scope := d.Get("scope").(string)
	artifactName := d.Get("name").(string)
	blueprintName := d.Get("blueprint_name").(string)

	resp, err := client.Delete(ctx, scope, blueprintName, artifactName)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error deleting policy artifact %s from blueprint %s in scope %s", artifactName, blueprintName, scope)
	}

	return nil
}

func flattenBlueprintPolicyAssignmentProperties(paap blueprint.PolicyAssignmentArtifactProperties) []interface{} {
	props := make(map[string]interface{})
	desc := paap.Description
	displayName := paap.DisplayName
	dependsOn := paap.DependsOn
	policyDefinitionID := paap.PolicyDefinitionID
	resourceGroup := paap.ResourceGroup
	params := flattenBlueprintPolicyAssignmentParameters(paap.Parameters)

	props["description"] = &desc
	props["display_name"] = &displayName
	props["depends_on"] = dependsOn
	props["policy_definition_id"] = &policyDefinitionID
	props["resource_group"] = resourceGroup
	props["parameters"] = params

	return []interface{}{props}
}

func flattenBlueprintPolicyAssignmentParameters(p map[string]*blueprint.ParameterValueBase) *schema.Set {
	params := &schema.Set{
		F: resourceBlueprintPolicyArtifactParametersHash,
	}

	for k, v := range p {
		param := make(map[string]*string)
		param[k] = v.Description
		params.Add(param)
	}

	return params
}

func resourceBlueprintPolicyArtifactParametersHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		if h, ok := m["description"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", h))
		}
	}

	return hashcode.String(buf.String())
}

func expandPolicyAssignmentArtifactProperties(d *schema.ResourceData) *blueprint.PolicyAssignmentArtifactProperties {
	propertiesRaw := d.Get("properties").([]interface{})
	properties := propertiesRaw[0].(map[string]interface{})

	displayName := properties["display_name"].(string)
	desc := properties["description"].(string)
	resourceGroup := properties["resource_group"].(string)
	policyDefinitionID := properties["policy_definition_id"].(string)
	paramsRaw := properties["parameters"].(*schema.Set).List()
	parameters := expandPolicyAssignmentArtifactParameters(paramsRaw)
	dependsOnRaw := properties["depends_on"].([]interface{})
	dependsOn := make([]string, 0)
	for k, v := range dependsOnRaw {
		dependsOn[k] = v.(string)
	}

	paap := blueprint.PolicyAssignmentArtifactProperties{
		DisplayName:        &displayName,
		Description:        &desc,
		DependsOn:          &dependsOn,
		PolicyDefinitionID: &policyDefinitionID,
		ResourceGroup:      &resourceGroup,
	}
	paap.Parameters = parameters

	return &paap
}

func expandPolicyAssignmentArtifactParameters(p []interface{}) map[string]*blueprint.ParameterValueBase {
	params := make(map[string]*blueprint.ParameterValueBase)
	for _, v := range p {
		param := v.(map[string]interface{})
		name := param["name"].(string)
		desc := param["description"].(string)
		params[name] = &blueprint.ParameterValueBase{
			Description: &desc,
		}
	}

	return params
}
