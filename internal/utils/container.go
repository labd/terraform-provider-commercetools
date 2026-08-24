package utils

import (
	"github.com/labd/commercetools-go-sdk/insights"
	"github.com/labd/commercetools-go-sdk/platform"
)

type ProviderData struct {
	Client *platform.ByProjectKeyRequestBuilder
	// InsightsClient talks to the Platform Insights API, which is a separate
	// API (and therefore a separate generated client) from the platform API.
	InsightsClient *insights.ByProjectKeyRequestBuilder
	Mutex          *MutexKV
}
