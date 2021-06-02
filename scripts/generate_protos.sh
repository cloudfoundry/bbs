DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$DIR/../models"
protoc --proto_path=$DIEGO_RELEASE_DIR/src/code.cloudfoundry.org/vendor:$DIEGO_RELEASE_DIR/src/code.cloudfoundry.org/vendor/github.com/golang/protobuf/ptypes/duration/:. --gogoslick_out=plugins=grpc:. *.proto
popd
