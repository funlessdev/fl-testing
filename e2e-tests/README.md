# End-to-end tests for the Funless serverless platform

To run all tests on an already deployed funless instance:
```
FL_TEST_HOST=<funless core address> go test -v ./test/
```

To run all tests, deploying a development funless instance to perform them:
```
FL_TEST_DEPLOY=true FL_TEST_HOST=http://localhost:4000 go test -v ./test/
```