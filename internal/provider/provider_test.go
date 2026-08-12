package provider

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
	"time"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"awsex": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

// testAccPreCheckLiveAws gates tests that create real AWS resources. It must be called
// before any direct AWS calls (e.g. testAccCreateCdn), which run before resource.Test
// gets a chance to apply its own TF_ACC gate.
func testAccPreCheckLiveAws(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC to run acceptance tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err == nil {
		_, err = cfg.Credentials.Retrieve(ctx)
	}
	if err != nil {
		t.Skipf("skipping: AWS credentials are not available (%v)", err)
	}
}
