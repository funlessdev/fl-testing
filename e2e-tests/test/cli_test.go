package tests

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/funlessdev/fl-testing/e2e-tests/internal/cli"
	"github.com/stretchr/testify/suite"
)

type SDKTestSuite struct {
	suite.Suite
	deploy      bool
	fnName      string
	fnNamespace string
	fnCode      *os.File
	fnHost      string
	fnArgs      interface{}
}

func (suite *SDKTestSuite) SetupSuite() {
	host := os.Getenv("FL_TEST_HOST")
	if host == "" {
		suite.T().Skip("set FL_TEST_HOST to run this test")
	}

	deploy := os.Getenv("FL_TEST_DEPLOY")
	if deploy != "" {
		suite.deploy, _ = strconv.ParseBool(deploy)
	} else {
		suite.deploy = false
	}

	suite.fnName = "hello-test"
	suite.fnNamespace = "helloNS"
	suite.fnArgs = map[string]string{"name": "Test"}
	source, err := os.OpenFile("../functions/hello.wasm", os.O_RDONLY, 0644)

	if err != nil {
		suite.T().Errorf("Error while reading source code file: %+v\n", err)
	}

	suite.fnCode = source
	suite.fnHost = host

	if suite.deploy == true {
		cli.RunFLCmd("admin", "dev")
		//wait for everything to be up
		time.Sleep(10 * time.Second)
	}
}

func (suite *SDKTestSuite) TearDownSuite() {
	if suite.deploy == true {
		cli.RunFLCmd("admin", "reset")
	}
}

func (suite *SDKTestSuite) TestOperationsSuccess() {
	// create function
	suite.Run("should successfully create function", func() {
		args := []string{"fn", "create", suite.fnName, "--source-file", "../functions/hello.wasm", "--namespace", suite.fnNamespace, "--no-build", "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.Equal(suite.fnName+"\n", result)
	})

	// invoke function
	suite.Run("should return the correct result when invoking hello", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnNamespace, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		suite.Equal("{\"payload\":\"Hello Test!\"}\n", result)
	})

	// delete function
	suite.Run("should successfully delete function", func() {
		args := []string{"fn", "delete", suite.fnName, "--namespace", suite.fnNamespace}
		result := cli.RunFLCmd(args...)
		suite.Equal(suite.fnName+"\n", result)
	})

	// create function with build
	suite.Run("should successfully build and create function", func() {
		fn := "built"
		ns := "ns"

		// Create
		args := []string{"fn", "create", fn, "--source-dir", "../functions/hello_rust", "--namespace", ns, "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(result, "fl: error"))

		// Invoke
		args = []string{"fn", "invoke", fn, "--namespace", ns, "-j", `{"name":"Build"}`}
		result = cli.RunFLCmd(args...)
		suite.Equal("{\"payload\":\"Hello Build!\"}\n", result)

		// Delete
		args = []string{"fn", "delete", fn, "--namespace", ns}
		result = cli.RunFLCmd(args...)
		suite.Equal(fn+"\n", result)
	})
}

func (suite *SDKTestSuite) TestOperationsFailure() {
	// invocation before creation
	suite.Run("should return an error when invoking a function before creating it", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnNamespace, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		suite.Equal("fl: error: Failed to invoke function: not found in given namespace", result)
	})

	// invocation on wrong namespace
	suite.Run("should return an error when invoking a function in the wrong namespace", func() {

		// Create the function first
		args := []string{"fn", "create", suite.fnName, "--source-file", "../functions/hello.wasm", "--namespace", suite.fnNamespace, "--no-build", "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.Equal(suite.fnName+"\n", result)

		// Invoke the function with a wrong namespace
		args = []string{"fn", "invoke", suite.fnName, "--namespace", "WRONG", "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		suite.Equal("fl: error: Failed to invoke function: not found in given namespace", result)

	})

	// invocation after deletion
	suite.Run("should return an error when invoking a function after deleting it", func() {
		// Create the function first
		args := []string{"fn", "create", suite.fnName, "--source-file", "../functions/hello.wasm", "--namespace", suite.fnNamespace, "--no-build", "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.Equal(suite.fnName+"\n", result)

		// Then delete it
		args = []string{"fn", "delete", suite.fnName, "--namespace", suite.fnNamespace}
		result = cli.RunFLCmd(args...)
		suite.Equal(suite.fnName+"\n", result)

		// Invoke the function
		args = []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnNamespace, "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		suite.Equal("fl: error: Failed to invoke function: not found in given namespace", result)
	})

	// delete before creation
	suite.Run("should return an error when deleting a function before creating it", func() {
		args := []string{"fn", "delete", suite.fnName, "--namespace", suite.fnNamespace}
		result := cli.RunFLCmd(args...)
		suite.Equal("fl: error: Failed to delete function: not found", result)
	})
}

func TestSDKSuite(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}
