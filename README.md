# BBS Server

Internal API to access the database for Diego.

## Code Generation

We generate code from the .proto (protobuf) files. We also generate a set of fakes from the interfaces we have.
To do so, just use `go generate`.

```
go generate ./...
```
