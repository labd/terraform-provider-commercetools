# commercetools Terraform Provider

[![Test status](https://github.com/labd/terraform-provider-commercetools/workflows/Run%20Tests/badge.svg)](https://github.com/labd/terraform-provider-commercetools/actions?query=workflow%3A%22Run+Tests%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/labd/terraform-provider-commercetools)](https://goreportcard.com/report/github.com/labd/terraform-provider-commercetools)

The Terraform commercetools provider allows you to configure
your [commercetools](https://commercetools.com/) project with
infrastructure-as-code principles.

# Commercial support

Need support implementing this terraform provider in your organization? Or are you missing features 
that need to be added, then we are able to offer support. Please contact us at opensource@labdigital.nl

# Quick start

[Read our documentation](https://registry.terraform.io/providers/labd/commercetools/latest/docs)
and check out the [examples](https://registry.terraform.io/providers/labd/commercetools/latest/docs/guides/extensions).

## Usage

The provider is distributed via the Terraform registry. To use it you need to configure
the [`required_provider`](https://www.terraform.io/language/providers/requirements#requiring-providers) block. For example:

```hcl
terraform {
  required_providers {
    commercetools = {
      source = "labd/commercetools"

      # It's recommended to pin the version, e.g.:
      # version = "~> 1.4.0"
    }
  }
}
```

# Binaries

Packages of the releases are available at
https://github.com/labd/terraform-provider-commercetools/releases See the
[terraform documentation](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins)
for more information about installing third-party providers.

# Contributing

## Requirements

- [Go](https://golang.org/doc/install) >= 1.20
- [Task](https://taskfile.dev/installation/)
- [Knope](https://knope.tech/installation/) (optional, for `knope document-change`)

## Building the provider

Clone repository to: `$GOPATH/src/github.com/labd/terraform-provider-commercetools`

```sh
$ mkdir -p $GOPATH/src/github.com/labd; cd $GOPATH/src/github.com/labd
$ git clone git@github.com:labd/terraform-provider-commercetools
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/labd/terraform-provider-commercetools
$ task build-local
```

A build is created `terraform-provider-commercetools_99.0.0` in the root directory and added to plugin folder available
locally:

```sh
~/.terraform.d/plugins/local/labd/commercetools/99.0.0/${OS_ARCH}/terraform-provider-commercetools_v99.0.0
```

Use version `99.0.0` in the provider to test your changes locally

```hcl
terraform {
  required_providers {
    commercetools = {
      source  = "labd/commercetools"
      version = "99.0.0"
    }
  }
}
```

## Adding new resources

When commercetools releases new features which include new resources these need to be implemented
in terraform as new resources too. This provider is currently undergoing a migration path to
move from [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk/) to the new
[terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework/). All new
resources therefore need to be written using the terraform-plugin-framework module. See the
`subscription` component as example.

## Debugging / Troubleshooting

There are two environment settings for troubleshooting:

- `TF_LOG=INFO` enables debug output for Terraform.
- `CTP_DEBUG=1` enables debug output for the Commercetools GO SDK this provider uses.

Note this generates a lot of output!

## Changelog and releasing

Every PR with a user facing change needs a change file in the `.changeset/` folder. Releases and
`CHANGELOG.md` are generated from these files by [Knope](https://knope.tech). Commit messages are not
used for the changelog.

Create one interactively by running this in the root of the project:

```sh
knope document-change
```

Or write the file by hand, for example `.changeset/fix-channel-address.md`:

```md
---
default: patch
---

# Fix `commercetools_channel` reading `address.additional_street_info`
```

The first heading is the changelog entry. Any text below it is added as extra detail. The type
(`default: ...`) selects the changelog section and the version bump:

| Type         | Section          | Version bump |
| ------------ | ---------------- | ------------ |
| `major`      | Breaking changes | major        |
| `minor`      | Added            | minor        |
| `patch`      | Fixed            | patch        |
| `deprecated` | Deprecated       | patch        |
| `security`   | Security         | patch        |
| `dependency` | Dependency       | patch        |

We recommend `patch` or `minor` for most changes. Dependabot PRs get a `dependency` change file
automatically.

On every push to `main` Knope opens or updates a `chore: prepare release x.y.z` PR that bumps the
version (kept as a comment on the module line in `go.mod`), updates `CHANGELOG.md` and removes the
change files. Merging that PR builds and signs the provider with GoReleaser and creates the `vx.y.z`
tag and GitHub release, which the Terraform Registry picks up.

## Testing

### Running the unit tests

```sh
$ task test
```

### Running the unit tests with coverage

```sh
$ task coverage
```

### Running an Acceptance Test

In order to run the full suite of Acceptance tests, run `task testacc`.

**NOTE:** Acceptance tests create real resources.

Prior to running the tests provider configuration details such as access keys
must be made available as environment variables.

Since we need to be able to create commercetools resources, we need the
commercetools API credentials. So in order for the acceptance tests to run
correctly please provide all the following:

```sh
export CTP_CLIENT_ID=...
export CTP_CLIENT_SECRET=...
export CTP_PROJECT_KEY=...
export CTP_SCOPES=...
```

For convenience, place a `testenv.sh` in your `local` folder (which is
included in .gitignore) where you can store these environment variables.

Tests can then be started by running

```sh
$ source local/testenv.sh
$ task testacc
```

## Authors

This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-commercetools)
for more information.
