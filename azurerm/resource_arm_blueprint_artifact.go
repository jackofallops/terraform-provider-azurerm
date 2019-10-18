package azurerm

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"net/http"
)

func resourceArmBlueprintArtifact() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmBlueprintArtifactCreateOrUpdate,
		Update: resourceArmBlueprintArtifactCreateOrUpdate,
		Delete: resourceArmBlueprintArtifactDelete,
		Read:   resourceArmBlueprintArtifactRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Optional: true,
				// todo Scope validation function
			},
			"blueprint_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(blueprint.KindArtifact),
					string(blueprint.KindPolicyAssignment),
					string(blueprint.KindRoleAssignment),
					string(blueprint.KindTemplate),
				}, true),
			},
			"template_artifact": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"policy_assignment_artifact", "role_assignment_artifact"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"depends_on": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"template": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.ValidateJsonString,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"policy_assignment_artifact": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"template_artifact"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"depends_on": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"policy_definition_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"role_assignment_artifact": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"template_artifact", "policy_assignment_artifact"},
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"depends_on": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"role_definition_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"principal_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							// Todo - Look at Elem properly here
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.NoEmptyStrings,
							},
						},
						"resource_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceArmBlueprintArtifactCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).blueprintArtifact
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	scope := d.Get("scope").(string)
	blueprintName := d.Get("blueprint_name").(string)

	kind := d.Get("kind")

	var artifact blueprint.BasicArtifact

	dName := d.Get("template_artifact.display_name").(string)
	desc := d.Get("template_artifact.description").(string)
	rg := d.Get("template_artifact.resource_group").(string)

	switch kind {
	case blueprint.KindTemplate:
		template := d.Get("template_artifact.template").(string)

		tArtifact := blueprint.TemplateArtifactProperties{
			DisplayName: &dName,
			Description: &desc,
			// TODO DependsOn
			Template:      template,
			ResourceGroup: &rg,
		}

		if params := expandParameters(d); params != nil {
			tArtifact.Parameters = params
		}

		artifact = blueprint.TemplateArtifact{
			Kind:                       blueprint.KindTemplate,
			TemplateArtifactProperties: &tArtifact,
		}

	case blueprint.KindRoleAssignment:
		// TODO Role assignment artifact case

	case blueprint.KindPolicyAssignment:
		// TODO Policy assignment artifact case
	}

	res, err := client.CreateOrUpdate(ctx, scope, blueprintName, name, artifact)
	if res.Response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error creating or updating blueprint artifact: %q", err)
	}

	return err
}

func expandParameters(d *schema.ResourceData) map[string]*blueprint.ParameterValueBase {
	params := d.Get("template_artifact.parameters").(map[string]string)
	if len(params) == 0 {
		return nil
	}
	ret := map[string]*blueprint.ParameterValueBase{}
	for k, v := range params {
		ret[k] = &blueprint.ParameterValueBase{Description: &v}
	}
	return ret
}

func resourceArmBlueprintArtifactRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceArmBlueprintArtifactDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
