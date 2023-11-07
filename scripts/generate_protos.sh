DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$DIR/../models"
protoc --proto_path=../../vendor:../../vendor/github.com/golang/protobuf/ptypes/duration/:. --gogoslick_out=plugins=grpc:. *.proto
popd
