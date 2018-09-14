# commercetools Terraform Provider

[![Travis Build Status](https://travis-ci.org/labd/terraform-provider-commercetools.svg?branch=master)](https://travis-ci.org/labd/terraform-provider-commercetools)
[![codecov](https://codecov.io/gh/LabD/terraform-provider-commercetools/branch/master/graph/badge.svg)](https://codecov.io/gh/LabD/terraform-provider-commercetools)
[![Go Report Card](https://goreportcard.com/badge/github.com/labd/terraform-provider-commercetools)](https://goreportcard.com/report/github.com/labd/terraform-provider-commercetools)
[![Documentation Status](https://readthedocs.org/projects/commercetools-terraform-provider/badge/?version=latest)](https://commercetools-terraform-provider.readthedocs.io/en/latest/?badge=latest)

# Status

This is the Terraform provider for commercetools. It allows you to configure your
commercetools project with infrastructure-as-code principles. The project is
still in early development and it doesn't support the complete commercetools
API yet, but it 'production' ready.

# Using the provider

[Read our documentation](https://readthedocs.org/projects/commercetools-terraform-provider) and check out the [examples](https://commercetools-terraform-provider.readthedocs.io/en/latest/examples/).

# Contributing

## Building the provider

Clone repository to: `$GOPATH/src/github.com/labd/terraform-provider-commercetools`

```sh
$ mkdir -p $GOPATH/src/github.com/labd; cd $GOPATH/src/github.com/labd
$ git clone git@github.com:labd/terraform-provider-commercetools
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/labd/terraform-provider-commercetools
$ make build
```

## Testing

### Running the unit tests

```sh
$ make test
```

### Running an Acceptance Test

In order to run the full suite of Acceptance tests, run `make testacc`.

**NOTE:** Acceptance tests create real resources.

Prior to running the tests provider configuration details such as access keys
must be made available as environment variables.

Since we need to be able to create commercetools resources, we need the
commercetools API credentials. So in order for the acceptance tests to run
correctly please provide all of the following:

```sh
export CTP_CLIENT_ID=...
export CTP_CLIENT_SECRET=...
export CTP_PROJECT_KEY=...
```
For convenience, place a `testenv.sh` in your `local` folder (which is
included in .gitignore) where you can store these environment variables.

Tests can then be started by running

```sh
$ source local/testenv.sh
$ make testacc
```

## Authors
This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-commercetools)
for more information.
