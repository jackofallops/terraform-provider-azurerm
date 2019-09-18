package blueprint

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/blueprint/mgmt/2018-11-01-preview/blueprint"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	BlueprintsClient  blueprint.BlueprintsClient
	AssignmentsClient blueprint.AssignmentsClient
}

func BuildClient(o *common.ClientOptions) *Client {
	c := Client{}

	c.BlueprintsClient = blueprint.NewBlueprintsClient()
	o.ConfigureClient(&c.BlueprintsClient.Client, o.ResourceManagerAuthorizer)

	c.AssignmentsClient = blueprint.NewAssignmentsClient()
	o.ConfigureClient(&c.AssignmentsClient.Client, o.ResourceManagerAuthorizer)

	return &c
}
