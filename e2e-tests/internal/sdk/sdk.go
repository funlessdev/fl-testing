package sdk

import (
	swagger "github.com/funlessdev/fl-client-sdk-go"
)

func BuildClient(host string) interface{} {
	apiConfig := swagger.NewConfiguration()
	apiConfig.BasePath = host
	apiClient := swagger.NewAPIClient(apiConfig)

	return apiClient
}
