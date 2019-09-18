package azurerm

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"net/http"
	"testing"
)

func TestAccAzureRMBlueprint_basic(t *testing.T) {
	resourceName := "azurerm_blueprint.test"
	ri := tf.AccRandTimeInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testCheckAzureBlueprintDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBlueprint_basic_subscription(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureBlueprintExists(resourceName),
					),
			},
		},
	})
}

func testCheckAzureBlueprintDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).blueprint
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_blueprint" {
			continue
		}
	name := rs.Primary.Attributes["name"]
	scope := rs.Primary.Attributes["scope"]

	resp, err := conn.Get(ctx, scope, name)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNotFound {
		fmt.Errorf("Blueprint still exists\n%v", resp)
	}

	}
	return nil
}

func testAccAzureRMBlueprint_basic_subscription(ri int) string {
	return fmt.Sprintf(`
data "azurerm_subscription" "current" {}

resource "azurerm_blueprint" "test" {
  name        = "acctestbp-%d"
  scope       = data.azurerm_subscription.current.id
  description = "accTest blueprint %d"
`, ri, ri )
}

func testCheckAzureBlueprintExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		scope := rs.Primary.Attributes["scope"]

		conn := testAccProvider.Meta().(*ArmClient).blueprint
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := conn.Get(ctx, scope, name)

		if err != nil {
			return fmt.Errorf("Bad: Get on blueprintClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: blueprint %q (scope: %q) does not exist", name, scope)
		}
		return nil
	}

}