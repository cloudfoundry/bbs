# BBS Server [![GoDoc](https://godoc.org/github.com/cloudfoundry/bbs?status.svg)](https://godoc.org/github.com/cloudfoundry/bbs)

**Note**: This repository should be imported as `code.cloudfoundry.org/bbs`.

API to access the database for Diego.

A general overview of the BBS is documented [here](doc).

## API

To interact with the BBS from outside of Diego, use the methods provided on the
ExternalClient interface, documented [here](https://godoc.org/github.com/cloudfoundry/bbs#ExternalClient).

Components within Diego may use the full [Client interface](https://godoc.org/github.com/cloudfoundry/bbs#Client) to modify internal state.

## Code Generation

You need the 3.0 version of the `protoc` compiler. If you're a Homebrew user
on Mac OS X, you get that by running:

```
brew install protobuf --devel
```

> If you already have an older version of protobuf installed, you will have to
> uninstall it first: `brew uninstall protobuf`

You also need the `gogoproto` compiler in you path:

```
go install github.com/gogo/protobuf/protoc-gen-gogoslick
```

We generate code from the .proto (protobuf) files. We also generate a set of
fakes from the interfaces we have.
To do so, just use `go generate`.

```
go generate ./...
```

## SQL

See the instructions in [Running the Experimental SQL Unit Tests](https://github.com/cloudfoundry/diego-release/blob/develop/CONTRIBUTING.md#running-the-experimental-sql-unit-tests)
for testing against a SQL backend
