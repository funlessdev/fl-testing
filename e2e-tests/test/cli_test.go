package tests

import (
	"fmt"
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
	deploy bool
	fnName string
	fnMod  string
	fnCode *os.File
	fnHost string
	fnArgs interface{}
}

func (suite *SDKTestSuite) SetupSuite() {
	host := os.Getenv("HOST")
	if host == "" {
		suite.T().Skip("set HOST to run this test, can also set DEPLOY=true to deploy funless")
	}

	deploy := os.Getenv("DEPLOY")
	if deploy != "" {
		suite.deploy, _ = strconv.ParseBool(deploy)
	} else {
		suite.deploy = false
	}

	suite.fnName = "hello_test"
	// suite.fnMod = "helloNS"
	suite.fnMod = "_"
	suite.fnArgs = map[string]string{"name": "Test"}
	source, err := os.OpenFile("../functions/hello.wasm", os.O_RDONLY, 0644)

	if err != nil {
		suite.T().Errorf("Error while reading source code file: %+v\n", err)
	}

	suite.fnCode = source
	suite.fnHost = host

	if suite.deploy == true {
		cli.RunFLCmd("admin", "deploy", "docker", "up")
		//wait for everything to be up
		time.Sleep(10 * time.Second)
	}
}

func (suite *SDKTestSuite) TearDownSuite() {
	if suite.deploy == true {
		cli.RunFLCmd("admin", "deploy", "docker", "down")
	}
}

//TODO 1: all 404 errors are rendered as "Not Found" => should be more explicative (issue with either CLI or funless, probably CLI error extraction)

//TODO 3: test modules
//TODO 3.1: test module create -> create function in module -> invoke function there
//TODO 3.2: test module delete -> all functions in the module are deleted (try `fn invoke <func> -n <mod>` after deleting the module)

func (suite *SDKTestSuite) TestOperationsSuccess() {
	// upload function
	suite.Run("should successfully create function", func() {
		args := []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))
	})

	// invoke function
	suite.Run("should return the correct result when invoking hello", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnMod, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		suite.Equal("{\"payload\":\"Hello Test!\"}\n", result)
	})

	// delete function
	suite.Run("should successfully delete function", func() {
		args := []string{"fn", "delete", suite.fnName, "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("\nSuccessfully deleted function %s/%s.", suite.fnMod, suite.fnName), result)
	})

	// create function
	suite.Run("should successfully build and create function", func() {
		fn := "built"
		ns := "_"

		// Create
		args := []string{"fn", "create", fn, "../functions/hello_rust", "--namespace", ns, "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// Invoke
		args = []string{"fn", "invoke", fn, "--namespace", ns, "-j", `{"name":"Build"}`}
		result = cli.RunFLCmd(args...)
		suite.Equal("{\"payload\":\"Hello Build!\"}\n", result)

		// Delete
		args = []string{"fn", "delete", fn, "--namespace", ns}
		result = cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("\nSuccessfully deleted function %s/%s.", ns, fn), result)
	})
	// build function
	suite.Run("should successfully build function", func() {
		fn := "built"
		// Build
		args := []string{"fn", "build", fn, "../functions/hello_rust", "--language", "rust"}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		suite.FileExists(fn + ".wasm")
		os.Remove(fn + ".wasm")
	})

}

func (suite *SDKTestSuite) TestOperationsFailure() {
	// invocation before creation
	suite.Run("should return an error when invoking a function before creating it", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnMod, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		//NOTE: see TODO no.1
		// suite.Equal("fl: error: Failed to invoke function: not found in given namespace", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))
	})

	// invocation on wrong namespace
	suite.Run("should return an error when invoking a function in the wrong namespace", func() {
		// Create the function first
		args := []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// Invoke the function with a wrong namespace
		args = []string{"fn", "invoke", suite.fnName, "--namespace", "WRONG", "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		//NOTE: see TODO no.1
		// suite.Equal("fl: error: Failed to invoke function: not found in given namespace", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))

		args = []string{"fn", "delete", suite.fnName, "--namespace", suite.fnMod}
		_ = cli.RunFLCmd(args...)
	})

	// invocation after deletion
	suite.Run("should return an error when invoking a function after deleting it", func() {
		// Create the function first
		args := []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// Then delete it
		args = []string{"fn", "delete", suite.fnName, "--namespace", suite.fnMod}
		result = cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("\nSuccessfully deleted function %s/%s.", suite.fnMod, suite.fnName), result)

		// Invoke the function
		args = []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnMod, "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		//NOTE: see TODO no.1
		// suite.Equal("fl: error: Failed to invoke function: not found in given namespace", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))
	})

	// delete before creation
	suite.Run("should return an error when deleting a function before creating it", func() {
		args := []string{"fn", "delete", suite.fnName, "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		//NOTE: see TODO no.1
		// suite.Equal("fl: error: Failed to delete function: not found", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))
	})
}

func TestSDKSuite(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}

func lastLine(result string) string {
	lines := strings.Split(result, "\n")
	return lines[len(lines)-1]
}
