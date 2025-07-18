package main

import (
	"terraform-provider-izysafe/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: provider.Provider})
}
