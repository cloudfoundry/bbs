---
title: Development Worklow (Code Generation with Protobuf)
expires_at: never
tags: [diego-release, bbs]
---

# Development Worklow (Code Generation with Protobuf)

The protobuf models in this repository require version 3.5 or later of the `protoc` compiler.

### OSX

On Mac OS X with [Homebrew](http://brew.sh/), run the following to install it:

```
brew install protobuf
```

### Linux

1. Download a zip archive of the latest protobuf release from [here](https://github.com/google/protobuf/releases).
1. Unzip the archive in `/usr/local` (including /bin and /include folders).
1. `chmod a+x /usr/local/bin/protoc` to make sure you can use the binary.

> If you already have an older version of protobuf installed, you must
> uninstall it first by running `brew uninstall protobuf`

Install the `gogoproto` compiler by running:

```
go install github.com/gogo/protobuf/protoc-gen-gogoslick
```

Run `go generate ./...` from the root directory of this repository to generate code from the `.proto` files as well as to generate fake implementations of certain interfaces for use in test code.

### Generating ruby models for BBS models

The following documentation assume the following versions:

1. [protoc](https://github.com/google/protobuf/releases) `> v3.5.0`
2. [ruby protobuf gem](https://github.com/ruby-protobuf/protobuf) `> 3.6.12`

Run the following commands from the `models` directory to generate `.pb.rb`
files for the BBS models:

1. `sed -i'' -e 's/package models/package diego.bbs.models/' ./*.proto`
1. `protoc -I../../vendor --proto_path=. --ruby_out=/path/to/ruby/files *.proto`

> [!NOTE]
> Replace `/path/to/ruby/files` with the desired destination of the `.pb.rb` files. That directory must exist before running this command.
