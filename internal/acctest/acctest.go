package acctest

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/labd/terraform-provider-commercetools/internal/provider"
)

var ProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error)
var Provider tfprotov5.ProviderServer

func init() {
	ProtoV5ProviderFactories = protoV5ProviderFactoriesInit("commercetools")
	Provider = providerserver.NewProtocol5(provider.New("testing"))()
	if err := ConfigureProvider(Provider); err != nil {
		panic(err)
	}
}

func protoV5ProviderFactoriesInit(providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	factories := make(map[string]func() (tfprotov5.ProviderServer, error), len(providerNames))

	for _, name := range providerNames {
		factories[name] = func() (tfprotov5.ProviderServer, error) {
			p := providerserver.NewProtocol5(provider.New("testing"))()
			if err := ConfigureProvider(p); err != nil {
				panic(err)
			}
			return p, nil
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
		"client_id":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"client_secret": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"project_key":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"scopes":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"api_url":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"token_url":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
