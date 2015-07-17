#!/bin/bash

pushd `dirname $0`
  pushd ../models
    protoc \
      --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. \
      --gogofast_out=. *.proto
  popd
popd
