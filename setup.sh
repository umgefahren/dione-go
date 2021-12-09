# shellcheck disable=SC2046
protoc -I=$(pwd)/protos --go_out=$(pwd) $(pwd)/protos/messages.proto