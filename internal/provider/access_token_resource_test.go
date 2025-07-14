package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTokenBasic(t *testing.T) {
	testAccPreCheckToken(t)
	random := acctest.RandString(6)
	randomNum := acctest.RandIntRange(0, 999999)
	randomPin := fmt.Sprintf("%06d", randomNum)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTokenConfigBasic(fmt.Sprintf("token-%s", random), randomPin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("ysafe_access_token.token-%s", random), "label", fmt.Sprintf("token-%s", random)),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("ysafe_access_token.token-%s", random), "token"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_token.token-%s", random), "expiry"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_token.token-%s", random), "operations"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_token.token-%s", random), "folders_allowed"),
					resource.TestCheckResourceAttr(fmt.Sprintf("ysafe_access_token.token-%s", random), "pin", randomPin),
				),
			},
		},
	})
}

func testAccTokenConfigBasic(name string, pin string) string {
	return fmt.Sprintf(
		`
		provider "izysafe" {
			token    = "%s"
			pin      = "%s"
		}

		resource "ysafe_access_token" "%s" {
			label = "%s"
			pin   = "%s"
		}
	`, os.Getenv("YSAFE_TOKEN"), os.Getenv("YSAFE_PIN"), name, name, pin)
}

func TestAccPinResource_optionalAttributes(t *testing.T) {

	resourceName := "ysafe_access_token.test"
	tests := []struct {
		name   string
		config string
		check  resource.TestCheckFunc
	}{
		{
			name:   "Expiry only",
			config: testAccPinResourceConfigWithExpiry(fmt.Sprintf("token%s", acctest.RandStringFromCharSet(6, acctest.CharSetAlpha)), fmt.Sprintf("%06d", acctest.RandIntRange(0, 999999)), fmt.Sprintf("%06d", acctest.RandIntRange(0, 999999))),
			check:  resource.TestCheckResourceAttrSet(resourceName, "expiry"),
		},
	}

	for _, tt := range tests {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheckToken(t) },
			ProviderFactories: testAccProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: tt.config,
					Check:  tt.check,
				},
			},
		})
	}
}

func testAccPinResourceBase(label, pin string) string {
	return fmt.Sprintf(`
	provider "izysafe" {
		token    = "%s"
		pin      = "%s"
	}

	resource "ysafe_access_token" "test" {
		label = "%s"
		pin   = "%s"
`, os.Getenv("YSAFE_TOKEN"), os.Getenv("YSAFE_PIN"), label, pin)
}

func testAccPinResourceConfigWithExpiry(label, pin, expiry string) string {
	base := testAccPinResourceBase(label, pin)
	return base + fmt.Sprintf(`  expiry = "%s"
}
`, expiry)
}
