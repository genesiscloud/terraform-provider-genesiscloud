# Terraform Provider for Genesis Cloud

The Genesis Cloud provider is used to interact with resources supported by [Genesis Cloud](https://www.genesiscloud.com/). The provider needs to be configured with the proper credentials before it can be used.

- [Documentation](https://www.terraform.io/docs/providers/genesiscloud/index.html)
- [Genesis Cloud API Documentation](https://developers.genesiscloud.com/)
- [How to generate an API key?](https://support.genesiscloud.com/support/solutions/articles/47001126146-how-to-generate-an-api-token-)
- [Terraform Website](https://www.terraform.io)

This repository is licensed under Mozilla Public License 2.0 (no copyleft exception) (see [LICENSE.txt](./LICENSE.txt)) and includes third-party code subject to third-party notices (see [THIRD-PARTY-NOTICES.txt](./THIRD-PARTY-NOTICES.txt)).

# Maintainers

This provider plugin is maintained by Genesis Cloud. If you encounter any problems, feel free to create issues or pull requests.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

## Using the provider

Consult the [provider documentation](docs/index.md).

## Development

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
