package utils

import (
	"github.com/labd/commercetools-go-sdk/platform"
)

type ProviderData struct {
	Client *platform.ByProjectKeyRequestBuilder
	Mutex  *MutexKV
}
