# Contributing to Cloud Service Broker

The Cloud Service Broker team uses GitHub and accepts contributions via
[pull request](https://help.github.com/articles/using-pull-requests).

See the [docs](https://github.gwd.broadcom.net/TNZ/cloud-service-broker/tree/main/docs) for design notes and other helpful information on getting started.

## Contribution Workflow

1. Fork the repository
1. Check out `main` of cloud-service-broker
1. Create a feature branch (`git checkout -b better_csb`)
1. Make changes on your branch
1. Run unit and integration tests (`make run-tests`)
1. Make clear commit message using [conventional commits style](https://www.conventionalcommits.org/en/v1.0.0/#summary)
3. Push to your fork (`git push origin better_csb`)
4. Submit your PR

### PR Considerations
We favor pull requests with very small, single commits with a single purpose.

Your pull request is much more likely to be accepted if:
* Your pull request includes tests (unit and integration)
* Your pull request is small and focused.
* Your pull request has a clear message that conveys the intent of your change.
