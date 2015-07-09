# BBS Server

Internal API to access the database for Diego.

## Code Generation

You need the 3.0 version of the `protoc` compiler. On OSX you get that by running:

```
brew install protobuf --devel
```

You also need the `gogoproto` compiler in you path:

```
go get github.com/gogo/protobuf/protoc-gen-gogofast
```

We generate code from the .proto (protobuf) files. We also generate a set of fakes from the interfaces we have.
To do so, just use `go generate`.

```
go generate ./...
```
