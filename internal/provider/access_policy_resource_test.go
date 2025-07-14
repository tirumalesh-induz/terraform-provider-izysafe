package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	policyAttrList = []string{"max_size", "max_file_size", "max_file_versions", "remove_older_versions", "default_ttl_for_files"}
)

func TestAccPolicyBasic(t *testing.T) {
	testAccPreCheckPolicy(t)
	random := acctest.RandString(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigBasic(fmt.Sprintf("proj_%s", random)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "name", fmt.Sprintf("proj_%s", random)),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "id"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "max_size"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "max_file_size"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "max_file_versions"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "remove_older_versions"),
					resource.TestCheckNoResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "default_ttl_for_files"),
				),
			},
		},
	})
}

func testAccPolicyConfigBasic(name string) string {
	return fmt.Sprintf(
		`
		provider "izysafe" {
		token    = "%s"
		pin      = "%s"
		}

		resource "ysafe_access_policy" "%s" {
		name = "%s"
		}
	`, os.Getenv("YSAFE_TOKEN"), os.Getenv("YSAFE_PIN"), name, name)
}

func TestAccPolicyOneAttribute(t *testing.T) {
	// Optional pre-flight check: bail out early if creds are missing.
	testAccPreCheckPolicy(t)

	// Random suffix to avoid name collisions between parallel tests.
	for idx := range len(policyAttrList) {

		random := acctest.RandString(6)
		randIdx := idx
		var randVal any
		if randIdx != 3 {
			randVal = acctest.RandIntRange(1, 10*1024*1024)
		} else {
			randVal = acctest.RandIntRange(0, 1)
			if randVal == 0 {
				randVal = true
			} else {
				randVal = false
			}
		}

		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "name", fmt.Sprintf("proj_%s", random)),
			resource.TestCheckResourceAttrSet(fmt.Sprintf("ysafe_access_policy.proj_%s", random), "id"),
		}

		for idx, attr := range policyAttrList {
			switch idx {
			case randIdx:
				checks = append(checks, resource.TestCheckResourceAttr(
					fmt.Sprintf("ysafe_access_policy.proj_%s", random),
					attr,
					fmt.Sprintf("%v", randVal),
				))
			case 3:
				checks = append(checks, resource.TestCheckResourceAttr(
					fmt.Sprintf("ysafe_access_policy.proj_%s", random),
					attr,
					fmt.Sprintf("%t", true),
				))
			default:
				checks = append(checks, resource.TestCheckNoResourceAttr(
					fmt.Sprintf("ysafe_access_policy.proj_%s", random),
					attr,
				))
			}
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckProjectDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccPolicyConfigOneAttribute(fmt.Sprintf("proj_%s", random), policyAttrList[randIdx], randVal),
					Check:  resource.ComposeTestCheckFunc(checks...),
				},
			},
		})
	}
}

func testAccPolicyConfigOneAttribute(name string, attr string, value any) string {
	return fmt.Sprintf(
		`
		provider "izysafe" {
			token    = "%s"
			pin      = "%s"
		}

		resource "ysafe_access_policy" "%s" {
			name = "%s"
			%s = %v
		}
	`, os.Getenv("YSAFE_TOKEN"), os.Getenv("YSAFE_PIN"), name, name, attr, value)
}
