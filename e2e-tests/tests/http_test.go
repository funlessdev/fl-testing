package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/funlessdev/fl-cli/pkg/deploy"
	"github.com/funlessdev/fl-testing/e2e-tests/internal/cli"
	"github.com/stretchr/testify/suite"
)

type HTTPTestSuite struct {
	suite.Suite
	ctx         context.Context
	deployer    deploy.DockerDeployer
	fnName      string
	fnNamespace string
	fnCode      string
	fnImage     string
	fnHost      string
	fnArgs      map[string]string
}

func (suite *HTTPTestSuite) SetupSuite() {
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

	deployer, err := cli.NewDeployer(suite.ctx)
	if err != nil {
		suite.T().Errorf("Error during docker deployer creation: %+v\n", err)
	}

	suite.deployer = deployer
	_ = cli.DeployDev(suite.ctx, suite.deployer)

	//wait for everything to be up
	time.Sleep(5 * time.Second)
}

func (suite *HTTPTestSuite) TearDownSuite() {
	_ = cli.DestroyDev(suite.ctx, suite.deployer)
}

func (suite *HTTPTestSuite) CreateFunction() (*http.Response, error) {
	createBody := map[string]string{"name": suite.fnName, "namespace": suite.fnNamespace, "code": suite.fnCode, "image": suite.fnImage}
	jsonBody, _ := json.Marshal(createBody)
	response, err := http.Post(suite.fnHost+"/create", "application/json", bytes.NewBuffer(jsonBody))
	return response, err
}

func (suite *HTTPTestSuite) DeleteFunction() (*http.Response, error) {
	deleteBody := map[string]string{"name": suite.fnName, "namespace": suite.fnNamespace}
	jsonBody, _ := json.Marshal(deleteBody)
	response, err := http.Post(suite.fnHost+"/delete", "application/json", bytes.NewBuffer(jsonBody))
	return response, err
}

func (suite *HTTPTestSuite) InvokeFunction() (*http.Response, error) {
	invokeBody := map[string]string{"function": suite.fnName, "namespace": suite.fnNamespace}
	jsonBody, _ := json.Marshal(invokeBody)
	response, err := http.Post(suite.fnHost+"/invoke", "application/json", bytes.NewBuffer(jsonBody))
	return response, err
}

func (suite *HTTPTestSuite) InvokeFunctionWithArgs(args string) (*http.Response, error) {
	invokeBody := fmt.Sprintf(`{"function": "%s", "namespace": "%s", "args": %s}`, suite.fnName, suite.fnNamespace, args)
	response, err := http.Post(suite.fnHost+"/invoke", "application/json", bytes.NewBuffer([]byte(invokeBody)))
	return response, err
}

func (suite *HTTPTestSuite) TestInvocationSuccess() {
	// create function
	suite.Run("should successfully create function", func() {
		response, err := suite.CreateFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)

		responseContent := string(responseBody)
		expectedResult := fmt.Sprintf(`{"result":"%s"}`, suite.fnName)
		suite.Equal(expectedResult, responseContent)
	})

	// invoke function
	suite.Run("should return no error when invoking an existing function", func() {
		response, err := suite.InvokeFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)
	})
	suite.Run("should return the correct result when invoking hellojs with no args", func() {
		response, err := suite.InvokeFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)

		responseContent := string(responseBody)
		expectedResult := `{"result":{"payload":"Hello World!"}}`
		suite.Equal(expectedResult, string(responseContent))

	})
	suite.Run("should return the correct result when invoking hellojs with args", func() {
		jsonArgs, _ := json.Marshal(suite.fnArgs)
		response, err := suite.InvokeFunctionWithArgs(string(jsonArgs))
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)

		responseContent := string(responseBody)
		expectedResult := fmt.Sprintf(`{"result":{"payload":"Hello %s!"}}`, suite.fnArgs["name"])
		suite.Equal(expectedResult, string(responseContent))
	})

	//delete function
	suite.Run("should successfully delete function", func() {
		response, err := suite.DeleteFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)

		responseContent := string(responseBody)
		expectedResult := fmt.Sprintf(`{"result":"%s"}`, suite.fnName)
		suite.Equal(expectedResult, responseContent)
	})
}

func (suite *HTTPTestSuite) TestInvocationFailure() {
	// invocation before creation
	suite.Run("should return an error when invoking a function before creating it", func() {
		response, err := suite.InvokeFunction()

		suite.NoError(err)
		suite.Equal(404, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)
		expectedResult := `{"error":"Failed to invoke function: function not found in given namespace"}`
		suite.Equal(expectedResult, string(responseBody))
	})

	// invocation on wrong namespace
	suite.Run("should return an error when invoking a function in the wrong namespace", func() {
		response, err := suite.CreateFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		invokeBody := map[string]string{"function": suite.fnName, "namespace": suite.fnNamespace + "_"}
		jsonBody, _ := json.Marshal(invokeBody)
		response, err = http.Post(suite.fnHost+"/invoke", "application/json", bytes.NewBuffer(jsonBody))

		suite.NoError(err)
		suite.Equal(404, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)
		expectedResult := `{"error":"Failed to invoke function: function not found in given namespace"}`
		suite.Equal(expectedResult, string(responseBody))
	})

	// invocation after deletion
	suite.Run("should return an error when invoking a function after deleting it", func() {
		response, err := suite.CreateFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		response, err = suite.DeleteFunction()
		suite.NoError(err)
		suite.Equal(200, response.StatusCode)

		response, err = suite.InvokeFunction()
		suite.NoError(err)
		suite.Equal(404, response.StatusCode)

		responseBody, err := io.ReadAll(response.Body)
		suite.NoError(err)
		expectedResult := `{"error":"Failed to invoke function: function not found in given namespace"}`
		suite.Equal(expectedResult, string(responseBody))
	})
}

func (suite *HTTPTestSuite) TestHTTPOnlyInvocationFailure() {
}

func TestHTTPSuite(t *testing.T) {
	suite.Run(t, new(HTTPTestSuite))
}
