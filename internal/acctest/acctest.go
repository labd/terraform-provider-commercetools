package acctest

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/commercetools"
	"github.com/labd/terraform-provider-commercetools/internal/provider"
)

var ProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)
var Provider tfprotov5.ProviderServer

func init() {
	if os.Getenv("TF_ACC") != "1" {
		log.Println("TF_ACC is not set, skipping acceptance tests")
		return
	}

	ProtoV5ProviderFactories = protoV5ProviderFactoriesInit("commercetools")
	newProvider := providerserver.NewProtocol5(provider.New("testing"))()
	if err := ConfigureProvider(newProvider); err != nil {
		panic(err)
	}

	// Init the old SDK provider
	requiredEnvs := []string{
		"CTP_CLIENT_ID",
		"CTP_CLIENT_SECRET",
		"CTP_PROJECT_KEY",
		"CTP_SCOPES",
		"CTP_API_URL",
		"CTP_AUTH_URL",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			log.Fatalf("%v must be set for acceptance tests", val)
		}
	}

	cfg := map[string]any{
		"client_id":     os.Getenv("CTP_CLIENT_ID"),
		"client_secret": os.Getenv("CTP_CLIENT_SECRET"),
		"project_key":   os.Getenv("CTP_PROJECT_KEY"),
		"scopes":        os.Getenv("CTP_SCOPES"),
		"api_url":       os.Getenv("CTP_API_URL"),
		"token_url":     os.Getenv("CTP_AUTH_URL"),
	}
	sdkProvider := commercetools.New("testing")()
	diags := sdkProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(cfg))
	if diags.HasError() {
		panic(diags)
	}

	// Mux the new and the old provider
	providers := []func() tfprotov5.ProviderServer{
		func() tfprotov5.ProviderServer { return newProvider },
		sdkProvider.GRPCProvider,
	}

	ctx := context.Background()
	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	Provider = muxServer

}

func protoV5ProviderFactoriesInit(providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	factories := make(map[string]func() (tfprotov5.ProviderServer, error), len(providerNames))

	for _, name := range providerNames {
		if name == "commercetools" {
			factories[name] = func() (tfprotov5.ProviderServer, error) {
				return Provider, nil
			}
		} else {
			panic("not implemented")
		}
	}

	return factories
}

func TestAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"CTP_CLIENT_ID",
		"CTP_CLIENT_SECRET",
		"CTP_PROJECT_KEY",
		"CTP_SCOPES",
		"CTP_API_URL",
		"CTP_AUTH_URL",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}
}

func ConfigureProvider(p tfprotov5.ProviderServer) error {
	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"client_id":     tftypes.String,
			"client_secret": tftypes.String,
			"project_key":   tftypes.String,
			"scopes":        tftypes.String,
			"api_url":       tftypes.String,
			"token_url":     tftypes.String,
		},
	}

	testValue := tftypes.NewValue(testType, map[string]tftypes.Value{
		"client_id":     tftypes.NewValue(tftypes.String, os.Getenv("CTP_CLIENT_ID")),
		"client_secret": tftypes.NewValue(tftypes.String, os.Getenv("CTP_CLIENT_SECRET")),
		"project_key":   tftypes.NewValue(tftypes.String, os.Getenv("CTP_PROJECT_KEY")),
		"scopes":        tftypes.NewValue(tftypes.String, os.Getenv("CTP_SCOPES")),
		"api_url":       tftypes.NewValue(tftypes.String, os.Getenv("CTP_API_URL")),
		"token_url":     tftypes.NewValue(tftypes.String, os.Getenv("CTP_AUTH_URL")),
	})

	testDynamicValue, err := tfprotov5.NewDynamicValue(testType, testValue)
	if err != nil {
		return err
	}

	_, err = p.ConfigureProvider(context.TODO(), &tfprotov5.ConfigureProviderRequest{
		TerraformVersion: "1.0.0",
		Config:           &testDynamicValue,
	})
	if err != nil {
		return err
	}
	return nil
}
