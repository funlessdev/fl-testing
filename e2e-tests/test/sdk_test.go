package tests

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/funlessdev/fl-cli/pkg/deploy"
	swagger "github.com/funlessdev/fl-client-sdk-go"
	"github.com/funlessdev/fl-testing/e2e-tests/internal/cli"
	"github.com/funlessdev/fl-testing/e2e-tests/internal/sdk"
	"github.com/stretchr/testify/suite"
)

type SDKTestSuite struct {
	suite.Suite
	ctx         context.Context
	deployer    deploy.DockerDeployer
	fnName      string
	fnNamespace string
	fnCode      string
	fnImage     string
	fnHost      string
	fnArgs      interface{}
	fnClient    *swagger.APIClient
}

func (suite *SDKTestSuite) SetupSuite() {
	host := os.Getenv("FL_TEST_HOST")
	if host == "" {
		suite.T().Skip("set FL_TEST_HOST to run this test")
	}

	suite.ctx = context.Background()
	suite.fnName = "hellojs"
	suite.fnNamespace = "helloNS"
	suite.fnImage = "nodejs"
	suite.fnArgs = map[string]string{"name": "Test"}
	source, err := os.ReadFile("../functions/hello.js")

	if err != nil {
		suite.T().Errorf("Error while reading source code file: %+v\n", err)
	}

	suite.fnCode = string(source)
	suite.fnHost = host

	suite.fnClient = sdk.BuildClient(suite.fnHost)

	deployer, err := cli.NewDeployer(suite.ctx)
	if err != nil {
		suite.T().Errorf("Error during docker deployer creation: %+v\n", err)
	}

	suite.deployer = deployer
	_ = cli.DeployDev(suite.ctx, suite.deployer)

	//wait for everything to be up
	time.Sleep(5 * time.Second)
}

func (suite *SDKTestSuite) TearDownSuite() {
	_ = cli.DestroyDev(suite.ctx, suite.deployer)
}

func (suite *SDKTestSuite) TestInvocationSuccess() {
	// create function
	suite.Run("should successfully create function", func() {
		result, _, err := suite.fnClient.DefaultApi.CreatePost(suite.ctx, swagger.FunctionCreation{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
			Code:      suite.fnCode,
			Image:     suite.fnImage,
		})
		suite.NoError(err)
		suite.Equal(suite.fnName, result.Result)
	})

	// invoke function
	suite.Run("should return no error when invoking an existing function", func() {
		_, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})

		suite.NoError(err)
	})
	suite.Run("should return the correct result when invoking hellojs with args", func() {
		result, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})
		suite.NoError(err)

		if result.Result != nil {
			name := suite.fnArgs.(map[string]string)["name"]
			decodedResult, jErr := json.Marshal(*result.Result)

			suite.NoError(jErr)
			suite.Equal(`{"payload":"Hello `+name+`!"}`, string(decodedResult))
		}

	})
	suite.Run("should return the correct result when invoking hellojs with no args", func() {
		result, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
		})
		suite.NoError(err)

		if result.Result != nil {
			decodedResult, jErr := json.Marshal(*result.Result)
			suite.NoError(jErr)
			suite.Equal(`{"payload":"Hello World!"}`, string(decodedResult))
		}
	})

	//delete function
	suite.Run("should successfully delete function", func() {
		result, _, err := suite.fnClient.DefaultApi.DeletePost(suite.ctx, swagger.FunctionDeletion{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
		})
		suite.NoError(err)
		suite.Equal(suite.fnName, result.Result)
	})
}

func (suite *SDKTestSuite) TestInvocationFailure() {
	// invocation before creation
	suite.Run("should return an error when invoking a function before creating it", func() {
		_, _, err := suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})

		suite.Error(err)
		suite.Equal("Failed to invoke function: function not found in given namespace", sdk.ExtractError(err).Error())
	})

	// invocation on wrong namespace
	suite.Run("should return an error when invoking a function in the wrong namespace", func() {
		_, _, err := suite.fnClient.DefaultApi.CreatePost(suite.ctx, swagger.FunctionCreation{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
			Code:      suite.fnCode,
			Image:     suite.fnImage,
		})
		suite.NoError(err)

		_, _, err = suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace + "_",
			Args:      &suite.fnArgs,
		})

		suite.Error(err)
		suite.Equal("Failed to invoke function: function not found in given namespace", sdk.ExtractError(err).Error())
	})

	// invocation after deletion
	suite.Run("should return an error when invoking a function after deleting it", func() {
		_, _, err := suite.fnClient.DefaultApi.CreatePost(suite.ctx, swagger.FunctionCreation{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
			Code:      suite.fnCode,
			Image:     suite.fnImage,
		})
		suite.NoError(err)

		_, _, err = suite.fnClient.DefaultApi.DeletePost(suite.ctx, swagger.FunctionDeletion{
			Name:      suite.fnName,
			Namespace: suite.fnNamespace,
		})
		suite.NoError(err)

		_, _, err = suite.fnClient.DefaultApi.InvokePost(suite.ctx, swagger.FunctionInvocation{
			Function:  suite.fnName,
			Namespace: suite.fnNamespace,
			Args:      &suite.fnArgs,
		})

		suite.Error(err)
		suite.Equal("Failed to invoke function: function not found in given namespace", sdk.ExtractError(err).Error())
	})
}

func TestSDKSuite(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}
