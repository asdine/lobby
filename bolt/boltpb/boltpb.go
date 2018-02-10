package boltpb

//go:generate protoc --go_out=. endpoint.proto
//go:generate protoc-go-inject-tag -input=./endpoint.pb.go
