package internal

//go:generate protoc --go_out=. item.proto bucket.proto
//go:generate protoc-go-inject-tag -input=./item.pb.go
//go:generate protoc-go-inject-tag -input=./bucket.pb.go
