package acctest

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labd/commercetools-go-sdk/checkout"
	"github.com/labd/commercetools-go-sdk/ctutils"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/labd/commercetools-go-sdk/platform"
	"golang.org/x/oauth2/clientcredentials"
)

func GetClient() (*platform.ByProjectKeyRequestBuilder, error) {
	clientID := os.Getenv("CTP_CLIENT_ID")
	clientSecret := os.Getenv("CTP_CLIENT_SECRET")
	projectKey := os.Getenv("CTP_PROJECT_KEY")
	authURL := os.Getenv("CTP_AUTH_URL")
	apiURL := os.Getenv("CTP_API_URL")
	scopesRaw := os.Getenv("CTP_SCOPES")

	oauthScopes := strings.Split(scopesRaw, " ")
	oauth2Config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       oauthScopes,
		TokenURL:     fmt.Sprintf("%s/oauth/token", authURL),
	}

	httpClient := &http.Client{
		Transport: ctutils.DebugTransport,
	}

	client, err := platform.NewClient(&platform.ClientConfig{
		URL:         apiURL,
		Credentials: oauth2Config,
		UserAgent:   "terraform-provider-commercetools/testing",
		HTTPClient:  httpClient,
	})
	if err != nil {
		return nil, err
	}

	return client.WithProjectKey(projectKey), nil
}

// GetCheckoutClient returns a client for the Checkout API, on CTP_CHECKOUT_URL
// or else on CTP_API_URL, as the mock serves both.
func GetCheckoutClient() (*checkout.ByProjectKeyRequestBuilder, error) {
	url := os.Getenv("CTP_CHECKOUT_URL")
	if url == "" {
		url = os.Getenv("CTP_API_URL")
	}
	client, err := checkout.NewClient(&checkout.ClientConfig{
		URL: url,
		Credentials: &clientcredentials.Config{
			ClientID:     os.Getenv("CTP_CLIENT_ID"),
			ClientSecret: os.Getenv("CTP_CLIENT_SECRET"),
			Scopes:       strings.Split(os.Getenv("CTP_SCOPES"), " "),
			TokenURL:     fmt.Sprintf("%s/oauth/token", os.Getenv("CTP_AUTH_URL")),
		},
		UserAgent:  "terraform-provider-commercetools/testing",
		HTTPClient: &http.Client{Transport: ctutils.DebugTransport},
	})
	if err != nil {
		return nil, err
	}
	return client.WithProjectKey(os.Getenv("CTP_PROJECT_KEY")), nil
}

func CheckApiResult(err error) error {
	if errors.Is(err, platform.ErrNotFound) {
		return nil
	}

	switch v := err.(type) {
	case platform.GenericRequestError:
		if v.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("unhandled error generic error returned (%d)", v.StatusCode)
	case platform.ResourceNotFoundError:
		return nil
	default:
		return fmt.Errorf("unexpected result returned")
	}
}

// integrationLayerURL is CTP_INTEGRATION_LAYER_URL, or else CTP_API_URL, as
// the mock serves both.
func integrationLayerURL() string {
	if url := os.Getenv("CTP_INTEGRATION_LAYER_URL"); url != "" {
		return url
	}
	return os.Getenv("CTP_API_URL")
}

// ProviderConfig is a provider block for another Project than CTP_PROJECT_KEY,
// with the credentials and hosts of the environment. The mock creates a
// Project when it is first addressed, so a test that changes Project-wide
// settings, such as those of the Commerce Integration Layer, can run beside
// the others on a Project of its own.
func ProviderConfig(projectKey string) string {
	return fmt.Sprintf(`
provider "commercetools" {
  client_id             = %q
  client_secret         = %q
  project_key           = %q
  scopes                = "manage_project:%s"
  api_url               = %q
  token_url             = %q
  integration_layer_url = %q
}
`, os.Getenv("CTP_CLIENT_ID"), os.Getenv("CTP_CLIENT_SECRET"), projectKey, projectKey,
		os.Getenv("CTP_API_URL"), os.Getenv("CTP_AUTH_URL"), integrationLayerURL())
}

// GetIntegrationLayerClient returns a client for the Commerce Integration
// Layer management API of the Project.
func GetIntegrationLayerClient(projectKey string) (*integrationlayer.ByProjectKeyRequestBuilder, error) {
	client, err := integrationlayer.NewClient(&integrationlayer.ClientConfig{
		URL: integrationLayerURL(),
		Credentials: &clientcredentials.Config{
			ClientID:     os.Getenv("CTP_CLIENT_ID"),
			ClientSecret: os.Getenv("CTP_CLIENT_SECRET"),
			Scopes:       []string{"manage_project:" + projectKey},
			TokenURL:     fmt.Sprintf("%s/oauth/token", os.Getenv("CTP_AUTH_URL")),
		},
		UserAgent:  "terraform-provider-commercetools/testing",
		HTTPClient: &http.Client{Transport: ctutils.DebugTransport},
	})
	if err != nil {
		return nil, err
	}
	return client.WithProjectKey(projectKey), nil
}
