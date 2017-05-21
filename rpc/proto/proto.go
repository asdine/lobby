package proto

//go:generate protoc --go_out=plugins=grpc:. bucket.proto registry.proto types.proto
//go:generate protoc-go-inject-tag -input=./types.pb.go
//go:generate protoc-go-inject-tag -input=./bucket.pb.go
//go:generate protoc-go-inject-tag -input=./registry.pb.go
