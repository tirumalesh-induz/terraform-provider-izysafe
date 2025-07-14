package provider

import (
	"terraform-provider-izysafe/internal/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The token data generated with the pin.",
			},
			"pin": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The six digit password created for authenticating user.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ysafe_access_token":  resourceAccessToken(),
			"ysafe_access_policy": resourceAccessPolicy(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	endpoint := "wss://files.ysafe.io:5577"
	pin := d.Get("pin").(string)

	client := client.GetClient(token, endpoint, pin)
	if client == nil {
		return nil, diag.Errorf("Failed to create client. Please check the token and pin. Contact support if the issue persists.")
	}
	return client, nil
}
