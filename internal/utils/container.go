package utils

import (
	"github.com/labd/commercetools-go-sdk/checkout"
	"github.com/labd/commercetools-go-sdk/connect"
	"github.com/labd/commercetools-go-sdk/insights"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/labd/commercetools-go-sdk/platform"
)

// IntegrationLayerLock serialises writes from the integration layer resources
// in one apply. Each route replaces only its own section, but every write bumps
// one version shared by all of them and none takes it as a precondition.
const IntegrationLayerLock = "commercetools-integration-layer"

type ProviderData struct {
	Client *platform.ByProjectKeyRequestBuilder
	// InsightsClient talks to the Platform Insights API, which is a separate
	// API (and therefore a separate generated client) from the platform API.
	InsightsClient *insights.ByProjectKeyRequestBuilder
	// IntegrationLayerClient talks to the Commerce Integration Layer management
	// API, on a host of its own. It is nil when no integration layer URL is
	// known.
	IntegrationLayerClient *integrationlayer.ByProjectKeyRequestBuilder
	// ConnectClient talks to the commercetools Connect API, on a host of its
	// own.
	ConnectClient *connect.ByProjectKeyRequestBuilder
	// ConnectorsClient talks to the Connect API's connectors, which belong to
	// the organization rather than to the Project.
	ConnectorsClient *connect.ConnectorsRequestBuilder
	// CheckoutClient talks to the commercetools Checkout API, on a host of
	// its own.
	CheckoutClient *checkout.ByProjectKeyRequestBuilder
	Mutex          *MutexKV
	// ProjectKey is the key of the Project the clients are for.
	ProjectKey string
}
