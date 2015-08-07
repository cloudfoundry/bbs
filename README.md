# BBS Server

Internal API to access the database for Diego.

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
go get github.com/gogo/protobuf/protoc-gen-gogoslick
```

We generate code from the .proto (protobuf) files. We also generate a set of
fakes from the interfaces we have.
To do so, just use `go generate`.

```
go generate ./...
```
