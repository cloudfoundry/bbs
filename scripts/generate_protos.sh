#! /bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
# regenerate protos for protoc plugin
pushd "${DIR}/../protoc-gen-go-bbs" > /dev/null
  protoc --proto_path=. --go_out=. --go_opt=paths=source_relative ./*.proto
  # we need the custom bbs code from the plugin because of references made by some of the model protos
  cp ./bbs.pb.go "${DIR}/../models/"

  # we also need to change the package after it's been copied away
  sed -i 's/package models/package main/g' ./bbs.pb.go
popd > /dev/null

# regenerate protos for models
pushd "${DIR}/../models" > /dev/null
  protoc --proto_path=.:../protoc-gen-go-bbs --go_out=. --go-bbs_out=. --go_opt=paths=source_relative --go-bbs_opt=paths=source_relative ./*.proto 
popd > /dev/null
