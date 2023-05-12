package acctest

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labd/commercetools-go-sdk/ctutils"
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
