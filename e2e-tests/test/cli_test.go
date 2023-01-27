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
	deploy   bool
	fnName   string
	fnMod    string
	fnNewMod string
	fnCode   *os.File
	fnHost   string
	fnArgs   interface{}
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
	suite.fnMod = "_"
	suite.fnNewMod = "test_mod"
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

//BUG 1: all 404 errors are rendered as "Not Found" => should be more explicative (issue with either CLI or funless, probably CLI error extraction)

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
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))
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
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))
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

	// create module
	suite.Run("should successfully create a module", func() {
		args := []string{"mod", "create", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("Successfully created module %s.\n", suite.fnNewMod), result)
	})
	// create function in module
	suite.Run("should successfully create function in a new module", func() {
		args := []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))
	})
	// invoke function in module
	suite.Run("should successfully invoke function in a new module", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnNewMod, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		suite.Equal("{\"payload\":\"Hello Test!\"}\n", result)
	})

	// list functions in module
	suite.Run("should successfully list functions in a module", func() {
		// create second function in module
		args := []string{"fn", "upload", suite.fnName + "2", "../functions/hello.wasm", "--namespace", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// list functions
		args = []string{"mod", "get", suite.fnNewMod}
		result = cli.RunFLCmd(args...)
		expected := fmt.Sprintf("Module: %s\nFunctions:\n%s\n%s\n", suite.fnNewMod, suite.fnName, suite.fnName+"2")
		suite.Equal(expected, result)

		// list functions with count
		args = []string{"mod", "get", suite.fnNewMod, "-c"}
		result = cli.RunFLCmd(args...)
		expected = fmt.Sprintf("%sCount: 2\n", expected)
		suite.Equal(expected, result)
	})

	// list modules
	suite.Run("should successfully list modules", func() {
		// list
		args := []string{"mod", "list"}
		result := cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("%s\n%s\n", suite.fnMod, suite.fnNewMod), result)
		// list with count
		args = []string{"mod", "list", "-c"}
		result = cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("%s\n%s\nCount: 2\n", suite.fnMod, suite.fnNewMod), result)
	})

	// delete module
	suite.Run("should successfully delete a module", func() {
		// delete
		args := []string{"mod", "delete", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("Successfully deleted module %s.\n", suite.fnNewMod), result)

		// check it was actually deleted
		args = []string{"mod", "list"}
		result = cli.RunFLCmd(args...)
		suite.Equal(fmt.Sprintf("%s\n", suite.fnMod), result)
		// BUG 2: see Issue #167 in funless repo
		/*
			args = []string{"mod", "get", suite.fnNewMod}
			result = cli.RunFLCmd(args...)
			suite.Equal("fl: error: Not Found", lastLine(result))
		*/
	})
}

func (suite *SDKTestSuite) TestOperationsFailure() {
	// invocation before creation
	suite.Run("should return an error when invoking a function before creating it", func() {
		args := []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnMod, "-j", `{"name":"Test"}`}
		result := cli.RunFLCmd(args...)
		//NOTE: see BUG 1
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
		//NOTE: see BUG 1
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
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// Invoke the function
		args = []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnMod, "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		//NOTE: see BUG 1
		// suite.Equal("fl: error: Failed to invoke function: not found in given namespace", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))
	})

	// delete before creation
	suite.Run("should return an error when deleting a function before creating it", func() {
		args := []string{"fn", "delete", suite.fnName, "--namespace", suite.fnMod}
		result := cli.RunFLCmd(args...)
		//NOTE: see BUG 1
		// suite.Equal("fl: error: Failed to delete function: not found", lastLine(result))
		suite.Equal("fl: error: Not Found", lastLine(result))
	})

	// create in nonexistent module
	suite.Run("should return an error when creating a function in a nonexistent model", func() {
		args := []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		//NOTE: see BUG 1
		suite.Equal("fl: error: Not Found", lastLine(result))
	})

	// invoke function after module was deleted
	suite.Run("should return an error when invoking a function after its module was deleted", func() {
		// create module
		args := []string{"mod", "create", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// create function
		args = []string{"fn", "upload", suite.fnName, "../functions/hello.wasm", "--namespace", suite.fnNewMod}
		result = cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// delete module
		args = []string{"mod", "delete", suite.fnNewMod}
		result = cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		// invoke function
		args = []string{"fn", "invoke", suite.fnName, "--namespace", suite.fnNewMod, "-j", `{"name":"Test"}`}
		result = cli.RunFLCmd(args...)
		//NOTE: see BUG 1
		suite.Equal("fl: error: Not Found", result)
	})
	// delete nonexistent module
	suite.Run("should return an error when trying to delete a nonexistent module", func() {
		args := []string{"mod", "delete", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.Equal("fl: error: Not Found", result)
	})
	// create existing module
	suite.Run("should return an error when trying to create an existing module", func() {
		args := []string{"mod", "create", suite.fnNewMod}
		result := cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))

		args = []string{"mod", "create", suite.fnNewMod}
		result = cli.RunFLCmd(args...)
		// BUG 3: the error returned when creating an entity that already exists does not have the correct shape (i.e. {"errors": {"details": ...}})
		suite.True(strings.HasPrefix(lastLine(result), "fl: error"))

		// cleanup
		args = []string{"mod", "delete", suite.fnNewMod}
		result = cli.RunFLCmd(args...)
		suite.False(strings.HasPrefix(lastLine(result), "fl: error"))
	})
}

func TestSDKSuite(t *testing.T) {
	suite.Run(t, new(SDKTestSuite))
}

func lastLine(result string) string {
	lines := strings.Split(result, "\n")
	return lines[len(lines)-1]
}
