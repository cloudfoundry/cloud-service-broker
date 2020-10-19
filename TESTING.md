# Testing Locally

## Unit Tests
We anticipate contributions to come in the form of Broker Packs so, more often than not, unit tests will not be necessary. 


In the unlikely event that you are editing the `Go` code the broker is written in, we ask that you use the standard `go test` framework before submitting a Pull Request.

## End to End Tests
End to end tests are generated from the documentation and examples and run outside the standard `go test` framework.
This ensures the auto-generated docs are always up-to-date and the examples work.
By executing the examples as an OSB client, it also ensures the service broker implements the OSB spec correctly.

To run the suite of end-to-end tests:

1. Start an instance of the broker `./cloud-service-broker serve`.
2. In a separate window, run the examples: `./cloud-service-broker client run-examples`
3. Wait for the examples to run and check the exit code. Exit codes other than 0 mean the end-to-end tests failed.

You can also target specific services in the end-to-end tests using the `--service-name` flag.
See `./cloud-service-broker client run-examples --help` for more details.

## Acceptance Testing

See [acceptance testing](acceptance-tests/README.md) for hints and tools for testing services.
