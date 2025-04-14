# Contributing Guide

This contributing guide is based on [CNCF template](https://contribute.cncf.io/maintainers/github/templates/required/contributing/#introduction)

* [New Contributor Guide](#contributing-guide)
  * [Ways to Contribute](#ways-to-contribute)
  * [Ask for Help](#ask-for-help)
  * [Development Environment Setup](#development-environment-setup)
  * [Sign Your Commits](#sign-your-commits)
  * [Pull Request Checklist](#pull-request-checklist)

  <!-- * [Pull Request Lifecycle](#pull-request-lifecycle) -->

Welcome! We are glad that you want to contribute to our project! 💖

As you get started, you are in the best position to give us feedback on areas of
our project that we need help with including:

* Problems found during setting up a new developer environment
* Gaps in our Quickstart Guide or documentation
* Bugs in our automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!

## Ways to Contribute

We welcome many different types of contributions including:

* New features
* Builds, CI/CD
* Bug fixes
* Documentation
* Testing

Almost everything happens through a GitHub pull request. Please see also [discussions](https://github.com/edgecomllc/eupf/discussions). 

<!--
## Find an Issue

[Instructions](https://contribute.cncf.io/maintainers/github/templates/required/contributing/#find-an-issue)

We have good first issues for new contributors and help wanted issues suitable
for any contributor. [good first issue](TODO) has extra information to
help you make your first contribution. [help wanted](TODO) are issues
suitable for someone who isn't a core maintainer and is good to move onto after
your first pull request.

Sometimes there won’t be any issues with these labels. That’s ok! There is
likely still something for you to work on. If you want to contribute but you
don’t know where to start or can't find a suitable issue, you can ⚠️ **explain how people can ask for an issue to work on**.

Once you see an issue that you'd like to work on, please post a comment saying
that you want to work on it. Something like "I want to work on this" is fine.
-->

## Ask for Help

The best way to reach us with a question when contributing is to ask on:

* The original github issue
* The discussion

<!--
## Pull Request Lifecycle

[Instructions](https://contribute.cncf.io/maintainers/github/templates/required/contributing/#pull-request-lifecycle)

⚠️ **Explain your pull request process**
-->


## Development Environment Setup

[Instructions](https://github.com/edgecomllc/eupf?tab=readme-ov-file#running-from-sources)

## Sign Your Commits

Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. We use the [Developer Certificate of Origin
(DCO)](https://probot.github.io/apps/dco/) as a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

You sign-off by adding the following to your commit messages. Your sign-off must
match the git user and email associated with the commit.

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running 

    git commit --amend -s 

## Pull Request Checklist

When you submit your pull request, or you push new commits to it, our automated
systems will run some checks on your new code. We require that your pull request
passes these checks, but we also have more criteria than just that before we can
accept and merge it. We recommend that you check the following things locally
before you submit your code:

- [x]  It passes tests: run the following command to run all of the tests locally: `go test -v ./cmd/...`
- [x]  Impacted code has new or updated tests
- [x]  Documentation created/updated
- [x]  All tests succeed when run by the CI build on a pull request before it is merged
- [x]  PR name and merge commit message are satisfy [Conventional Commits specification](https://www.conventionalcommits.org/) in order to keep [Release Notes](https://github.com/edgecomllc/eupf/releases) and main branch history clean

