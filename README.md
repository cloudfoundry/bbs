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

### Generating ruby models for bbs models

The following documentation assume the following versions:

1. [protoc](https://developers.google.com/protocol-buffers/docs/downloads) `> v3.0.0`
2. [ruby protobuf gem](https://github.com/ruby-protobuf/protobuf) `> 3.6.12`

Run the following commands from the `models` directory to generate `.pb.rb`
files for the bbs models:

1. `gem install protobuf`
2. cp `which protoc-gen-ruby`{,2}
3. protoc -I$GOPATH/src --proto_path=. --ruby2_out=/path/to/ruby/files *.proto

**Note** Replace `/path/to/ruby/files` with the desired destination of the
`.pb.rb` files. That directory must exist before running this command.

**Note** The above steps assume that
`github.com/gogo/protobuf/gogoproto/gogo.proto` is on the `GOPATH`.

**Note** Since protoc v3 now ships with a ruby generator, the built-in
generator will mask the gem's binary. This requires a small hack in order to be
able to use the protobuf gem, the hack is simply to rename the protobuf gem's
binary to be `ruby2` and ask protoc to generate `ruby2` code which will force
it to use the binary. For more information please
[read this github issue](https://github.com/ruby-protobuf/protobuf/issues/341)

## SQL

See the instructions in [Running the Experimental SQL Unit Tests](https://github.com/cloudfoundry/diego-release/blob/develop/CONTRIBUTING.md#running-the-experimental-sql-unit-tests)
for testing against a SQL backend
