package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/funlessdev/fl-cli/cmd/fl/app"
)

func RunFLCmd(args ...string) (out string) {
	cmd := append([]string{"fl"}, args...)
	os.Args = cmd

	// capture stdout to avoid too many cli prints, perhaps add a --quiet flag to fl-cli?
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		res, _ := ioutil.ReadAll(r)
		out = string(res)
		os.Stdout = rescueStdout
	}()

	// Capture stdout in w
	ctx, err := app.ParseCMD("ve2e-testing")
	if err != nil {
		return
	}
	if err = app.Run(ctx); err != nil {
		fmt.Print(cmd[0] + ": error: " + err.Error())
		return
	}

	return
}
