# End-to-end tests for the Funless serverless platform

To run all tests on an already deployed funless instance:
```
HOST=<funless core address> go test -v ./test/
```

To run all tests, deploying a development funless instance to perform them:
```
DEPLOY=true HOST=http://localhost:4000 go test -v ./test/
```