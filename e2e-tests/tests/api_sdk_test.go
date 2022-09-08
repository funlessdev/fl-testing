package tests

import (
	"context"
	"flag"
	"testing"
)

func TestAPISDK(t *testing.T) {
	testFn := "hellojs"
	testNs := "helloNS"
	testSource := "../functions/hello.js"
	testCtx := context.Background()
	host := flag.String("host", "http://localhost:4001", "Host protocol, address and port")

}
