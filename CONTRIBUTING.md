# Contributing to Cloud Service Broker

The Cloud Service Broker team uses GitHub and accepts contributions via
[pull request](https://help.github.com/articles/using-pull-requests).

See the [docs](https://github.com/cloudfoundry/cloud-service-broker/tree/main/docs) for design notes and other helpful information on getting started.

## Contributor License Agreement

Follow these steps to make a contribution to any of our open source repositories:

1. Ensure that you have signed our CLA Agreement [here](https://www.cloudfoundry.org/community/cla).
1. Set your name and email (these should match the information on your submitted CLA)

        git config --global user.name "Firstname Lastname"
        git config --global user.email "your_email@example.com"

1. All contributions must be sent using GitHub pull requests as they create a nice audit trail and structured approach.
   The originating github user has to either have a github id on-file with the list of approved users that have signed the CLA or they can be a public "member" of a GitHub organization for a group that has signed the corporate CLA. This enables the corporations to manage their users themselves instead of having to tell us when someone joins/leaves an organization. By removing a user from an organization's GitHub account, their new contributions are no longer approved because they are no longer covered under a CLA.

   If a contribution is deemed to be covered by an existing CLA, then it is analyzed for engineering quality and product fit before merging it.

   If a contribution is not covered by the CLA, then the automated CLA system notifies the submitter politely that we cannot identify their CLA and ask them to sign either an individual or corporate CLA. This happens automatically as a comment on pull requests.

   When the project receives a new CLA, it is recorded in the project records, the CLA is added to the database for the automated system uses, then we manually make the Pull Request as having a CLA on-file.

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

