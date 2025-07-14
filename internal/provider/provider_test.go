package provider_test

import (
	"testing"

	"terraform-provider-izysafe/internal/provider"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	p := provider.Provider()

	p.TerraformVersion = "1.0.0"

	err := p.InternalValidate()
	if err != nil {
		t.Fatal(err)
	}
}
