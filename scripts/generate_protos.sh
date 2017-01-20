if [[ "$OSTYPE" == "darwin"* ]]; then
  echo "OSX workstation detected, updating protobuf binaries..."
  rm $GOPATH/bin/protoc-gen-gogoslick
  go install -v github.com/gogo/protobuf/protoc-gen-gogoslick
fi

protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. --gogoslick_out=plugins=grpc:. *.proto
