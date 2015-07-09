package models

//go:generate protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. --gogofast_out=. actual_lrp.proto desired_lrp.proto modification_tag.proto actions.proto environment_variables.proto

const (
	maximumAnnotationLength = 10 * 1024
	maximumRouteLength      = 4 * 1024
)
