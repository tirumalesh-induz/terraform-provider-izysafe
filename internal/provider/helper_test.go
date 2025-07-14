package provider_test

import (
	"fmt"
	"os"
	"terraform-provider-izysafe/internal/client"
	"terraform-provider-izysafe/internal/provider"
	"testing"

	"terraform-provider-izysafe/internal/proto/request"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"ysafe": func() (*schema.Provider, error) {
		return provider.Provider(), nil
	},
}

func testAccPreCheckPolicy(t *testing.T) {
	if v := os.Getenv("YSAFE_TOKEN"); v == "" {
		t.Skip("YSAFE_TOKEN env var must be set for acceptance tests")
	}
	if v := os.Getenv("YSAFE_PIN"); v == "" {
		t.Skip("YSAFE_PIN env var must be set for acceptance tests")
	}
}

func testAccPreCheckToken(t *testing.T) {
	if v := os.Getenv("YSAFE_TOKEN"); v == "" {
		t.Skip("YSAFE_TOKEN env var must be set for acceptance tests")
	}
	if v := os.Getenv("YSAFE_PIN"); v == "" {
		t.Skip("YSAFE_PIN env var must be set for acceptance tests")
	}
}

func verifyDestroyFolder(name string, client *client.Client) bool {
	getMeta := request.GetMetaFromPath{
		Path:       "/" + name,
		Trashed:    false,
		TypeOfPath: 0,
	}
	req := request.Request{
		Operation: &request.Request_GetMetaFromPath{
			GetMetaFromPath: &getMeta,
		},
	}
	res, err := client.Send(&req)
	if err != nil {
		return true
	}
	if res == nil {
		return true
	}
	if res.GetGetMetaFromPath().Status == 0 {
		return true
	}
	return false
}

func verifyDestroyPin() bool {

	return false
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	// Iterate over resources left in state; confirm they are really gone
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ysafe_access_policy" && rs.Type != "ysafe_access_token" {
			continue
		}
		endpoint := "wss://files.ysafe.io:5577"
		pin := os.Getenv("YSAFE_PIN")
		token := os.Getenv("YSAFE_TOKEN")
		name := rs.Primary.ID
		client := client.GetClient(token, endpoint, pin)
		if verifyDestroyFolder(name, client) {
			return fmt.Errorf("resource %s not destroyed.", name)
		}
		if verifyDestroyPin() {

		}
	}
	return nil
}
