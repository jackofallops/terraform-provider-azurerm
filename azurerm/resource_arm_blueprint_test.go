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

func TestAccAzureRMBlueprint_basic(t *testing.T) {
	ri := tf.AccRandTimeInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureBlueprintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBlueprint_basic_subscription(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureBlueprintExists("azurerm_blueprint.test_subscription"),
				),
			},
			// Following test should have target_scope as `managementGroup` but fails an enum check in the API despite being in the spec
			// https://github.com/Azure/azure-rest-api-specs/blob/282efa7dd8301ba615d8741f740f1ed7f500fed1/specification/blueprint/resource-manager/Microsoft.Blueprint/preview/2018-11-01-preview/blueprintDefinition.json#L835
			{
				Config: testAccAzureRMBlueprint_basic_managementGroup(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureBlueprintExists("azurerm_blueprint.test_managementGroup"),
				),
			},
		},
	})
}

func testCheckAzureBlueprintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).Blueprint.BlueprintsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_blueprint" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		scope := rs.Primary.Attributes["scope"]

		resp, err := conn.Get(ctx, scope, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return err
			}
		}
	}
	return nil
}

func testCheckAzureBlueprintExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		scope := rs.Primary.Attributes["scope"]

		client := testAccProvider.Meta().(*ArmClient).Blueprint.BlueprintsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, scope, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on blueprintClient: %+v", err)
		}

		if resp.Response.Response.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: blueprint %q (scope: %q) does not exist", name, scope)
		}
		return nil
	}

}

func testAccAzureRMBlueprint_basic_subscription(ri int) string {
	return fmt.Sprintf(`
data "azurerm_client_config" "test" {}

resource "azurerm_blueprint" "test_subscription" {
  name  = "acctestbp-sub-%d"
  scope = join("",["/subscriptions/",data.azurerm_client_config.test.subscription_id])
  type  = "Microsoft.Blueprint/blueprints"
  properties {
    description  = "accTest blueprint %d"
    display_name = "accTest blueprint"
    target_scope = "subscription"
  }
}
`, ri, ri)
}

func testAccAzureRMBlueprint_basic_managementGroup(ri int) string {
	return fmt.Sprintf(`
data "azurerm_client_config" "test" {}

resource "azurerm_blueprint" "test_managementGroup" {
 name  = "acctestbp-mg-%d"
 scope = join("",["/providers/Microsoft.Management/managementGroups/",data.azurerm_client_config.test.tenant_id])
 type  = "Microsoft.Blueprint/blueprints"
 properties {
   description  = "accTest blueprint %d"
   display_name = "accTest blueprint"
   target_scope = "subscription"
 }
}
`, ri, ri)
}
