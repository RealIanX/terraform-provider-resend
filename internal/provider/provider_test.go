package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/realianx/terraform-provider-resend/internal/provider"
)

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("RESEND_API_KEY") == "" {
		t.Skip("RESEND_API_KEY must be set for acceptance tests")
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"resend": providerserver.NewProtocol6WithError(provider.New("test")()),
}
