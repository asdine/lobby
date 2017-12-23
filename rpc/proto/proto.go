package proto

//go:generate protoc --go_out=plugins=grpc:. topic.proto registry.proto
//go:generate protoc-go-inject-tag -input=./topic.pb.go
//go:generate protoc-go-inject-tag -input=./registry.pb.go
