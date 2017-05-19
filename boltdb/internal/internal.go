package internal

//go:generate protoc --go_out=. item.proto
//go:generate protoc-go-inject-tag -input=./item.pb.go
