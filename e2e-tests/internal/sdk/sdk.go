package sdk

import (
	"errors"

	swagger "github.com/funlessdev/fl-client-sdk-go"
)

func BuildClient(host string) *swagger.APIClient {
	apiConfig := swagger.NewConfiguration()
	apiConfig.BasePath = host
	apiClient := swagger.NewAPIClient(apiConfig)

	return apiClient
}

func ExtractError(err error) error {
	swaggerError, ok_sw := err.(swagger.GenericSwaggerError)
	if ok_sw {
		switch swaggerError.Model().(type) {
		case swagger.FunctionCreationError:
			specificError := swaggerError.Model().(swagger.FunctionCreationError)
			return errors.New(specificError.Error_)
		case swagger.FunctionDeletionError:
			specificError := swaggerError.Model().(swagger.FunctionDeletionError)
			return errors.New(specificError.Error_)
		case swagger.FunctionInvocationError:
			specificError := swaggerError.Model().(swagger.FunctionInvocationError)
			return errors.New(specificError.Error_)
		}
	}
	return err
}
