package azurerm

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"net/http"
	"testing"
)

func TestAccAzureRMBlueprintPolicyAssignment_basic(t *testing.T) {
	ri := tf.AccRandTimeInt()
	resourceName := ""
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMBlueprintPolicyAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBlueprintPolicyAssignment_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMBlueprintPolicyAssignmentExists(resourceName),
				),
			},
		},
	})
}

func testCheckAzureRMBlueprintPolicyAssignmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).Blueprint.ArtifactsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_blueprint_policy_assignment_artifact" {
			continue
		}
		name := rs.Primary.Attributes["name"]
		scope := rs.Primary.Attributes["scope"]
		blueprintName := rs.Primary.Attributes["blueprint_name"]

		resp, err := conn.Get(ctx, scope, blueprintName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return err
			}
		}
	}
	return nil
}

func testCheckAzureRMBlueprintPolicyAssignmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		scope := rs.Primary.Attributes["scope"]
		blueprintName := rs.Primary.Attributes["blueprint_name"]

		conn := testAccProvider.Meta().(*ArmClient).Blueprint.ArtifactsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := conn.Get(ctx, scope, blueprintName, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on artifactClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: artifact %q in blueprint %q (scope: %q) does not exist", name, blueprintName, scope)
		}
		return nil
	}
}

func testAccAzureRMBlueprintPolicyAssignment_basic(ri int) string {
	return fmt.Sprintf(`
data "azurerm_client_config" "test" {}

resource "azurerm_blueprint" "test" {
  name  = "acctestbp-sub-%d"
  scope = join("",["/subscriptions/",data.azurerm_client_config.test.subscription_id])
  type  = "Microsoft.Blueprint/blueprints"
  properties {
    description  = "accTest blueprint %d"
    display_name = "accTest blueprint"
    target_scope = "subscription"
	resource_groups {
      name         = "accTest-rg"
      location     = "westeurope"
      display_name = "blueprints acceptance test resource group"
      description  = "blueprints acceptance test resource group full description"
	  tags         = {
        accTest = "true"
      }
	}
    parameters {
	  name           = "Policy_Allowed-VM-SKUs"
	  type           = "array"
	  display_name   = "Virtual Machine SKUs you want to ALLOW"
	  default_value  = base64encode(join(",", ["Standard_B2ms", "Standard_DS1_v2", "Standard_F2s_v2"]))
      description    = "Policy_Allowed-VM-SKUs"
	  allowed_values = [
          "Standard_A1_v2",
          "Standard_A2m_v2",
          "Standard_A2_v2",
          "Standard_A4m_v2",
          "Standard_A4_v2",
          "Standard_A8m_v2",
          "Standard_A8_v2",
          "Standard_B1ls",
          "Standard_B1ms",
          "Standard_B1s",
          "Standard_B2ms",
          "Standard_B2s",
          "Standard_B4ms",
          "Standard_B8ms",
          "Standard_D1_v2",
          "Standard_D2s_v3",
          "Standard_D2_v2",
          "Standard_D2_v3",
          "Standard_D3_v2",
          "Standard_D4s_v3",
          "Standard_D4_v2",
          "Standard_D4_v3",
          "Standard_D5_v2",
          "Standard_D8s_v3",
          "Standard_D8_v3",
          "Standard_D11_v2",
          "Standard_D12_v2",
          "Standard_D13_v2",
          "Standard_D14_v2",
          "Standard_D15_v2",
          "Standard_D16s_v3",
          "Standard_D16_v3",
          "Standard_D32s_v3",
          "Standard_D32_v3",
          "Standard_D64s_v3",
          "Standard_D64_v3",
          "Standard_DC2s",
          "Standard_DC4s",
          "Standard_DS1_v2",
          "Standard_DS2_v2",
          "Standard_DS3_v2",
          "Standard_DS4_v2",
          "Standard_DS5_v2",
          "Standard_DS11-1_v2",
          "Standard_DS11_v2",
          "Standard_DS12-1_v2",
          "Standard_DS12-2_v2",
          "Standard_DS12_v2",
          "Standard_DS13-2_v2",
          "Standard_DS13-4_v2",
          "Standard_DS13_v2",
          "Standard_DS14-4_v2",
          "Standard_DS14-8_v2",
          "Standard_DS14_v2",
          "Standard_DS15_v2",
          "Standard_E2s_v3",
          "Standard_E2_v3",
          "Standard_E4-2s_v3",
          "Standard_E4s_v3",
          "Standard_E4_v3",
          "Standard_E8-2s_v3",
          "Standard_E8-4s_v3",
          "Standard_E8s_v3",
          "Standard_E8_v3",
          "Standard_E16-4s_v3",
          "Standard_E16-8s_v3",
          "Standard_E16s_v3",
          "Standard_E16_v3",
          "Standard_E20s_v3",
          "Standard_E20_v3",
          "Standard_E32-8s_v3",
          "Standard_E32-16s_v3",
          "Standard_E32s_v3",
          "Standard_E32_v3",
          "Standard_E64-16s_v3",
          "Standard_E64-32s_v3",
          "Standard_E64is_v3",
          "Standard_E64i_v3",
          "Standard_E64s_v3",
          "Standard_E64_v3",
          "Standard_F1s",
          "Standard_F2s",
          "Standard_F2s_v2",
          "Standard_F4s",
          "Standard_F4s_v2",
          "Standard_F8s",
          "Standard_F8s_v2",
          "Standard_F16s",
          "Standard_F16s_v2",
          "Standard_F32s_v2",
          "Standard_F64s_v2",
          "Standard_F72s_v2",
          "Standard_GS1",
          "Standard_GS2",
          "Standard_GS3",
          "Standard_GS4",
          "Standard_GS4-4",
          "Standard_GS4-8",
          "Standard_GS5",
          "Standard_GS5-8",
          "Standard_GS5-16",
          "Standard_H8",
          "Standard_H8m",
          "Standard_H16",
          "Standard_H16m",
          "Standard_H16mr",
          "Standard_H16r",
          "Standard_HB60rs",
          "Standard_HC44rs",
          "Standard_L4s",
          "Standard_L8s",
          "Standard_L8s_v2",
          "Standard_L16s",
          "Standard_L16s_v2",
          "Standard_L32s",
          "Standard_L32s_v2",
          "Standard_L64s_v2",
          "Standard_L80s_v2",
          "Standard_M8-2ms",
          "Standard_M8-4ms",
          "Standard_M8ms",
          "Standard_M16-4ms",
          "Standard_M16-8ms",
          "Standard_M16ms",
          "Standard_M32-8ms",
          "Standard_M32-16ms",
          "Standard_M32ls",
          "Standard_M32ms",
          "Standard_M32ts",
          "Standard_M64",
          "Standard_M64-16ms",
          "Standard_M64-32ms",
          "Standard_M64ls",
          "Standard_M64m",
          "Standard_M64ms",
          "Standard_M64s",
          "Standard_M128",
          "Standard_M128-32ms",
          "Standard_M128-64ms",
          "Standard_M128m",
          "Standard_M128ms",
          "Standard_M128s",
          "Standard_NC6",
          "Standard_NC6s_v2",
          "Standard_NC6s_v3",
          "Standard_NC12",
          "Standard_NC12s_v2",
          "Standard_NC12s_v3",
          "Standard_NC24",
          "Standard_NC24r",
          "Standard_NC24rs_v2",
          "Standard_NC24rs_v3",
          "Standard_NC24s_v2",
          "Standard_NC24s_v3",
          "Standard_ND6s",
          "Standard_ND12s",
          "Standard_ND24rs",
          "Standard_ND24s",
          "Standard_NV6",
          "Standard_NV6s_v2",
          "Standard_NV12",
          "Standard_NV12s_v2",
          "Standard_NV24",
          "Standard_NV24s_v2"
        ]
	}
  }
}

resource "azurerm_blueprint_policy_assignment_artifact" "test" {
  name = "accTest_policyArtifact-%d"
  scope = azurerm_blueprint.test.scope
  blueprint_name = azurerm_blueprint.test.name
  properties {
    policy_definition_id = "/providers/Microsoft.Authorization/policyDefinitions/cccc23c7-8427-4f53-ad12-b6a63eb452b3"
    display_name = "Allowed virtual machine SKUs"
    depends_on = []
    parameters {
      name = "listOfAllowedSKUs"
      description = "[parameters('Policy_Allowed-VM-SKUs')]"
    }
  }
}
`, ri, ri, ri)
}
