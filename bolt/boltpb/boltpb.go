package boltpb

//go:generate protoc --go_out=. message.proto topic.proto
//go:generate protoc-go-inject-tag -input=./topic.pb.go
//go:generate protoc-go-inject-tag -input=./message.pb.go
