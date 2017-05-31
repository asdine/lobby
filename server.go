package lobby

import "net"

// A Server serves incoming requests.
type Server interface {
	// Name returns the server short description. e.g: "http", "grpc", etc.
	Name() string

	// Serve incoming requests. Must block.
	Serve(net.Listener) error

	// Stop gracefully stops the server.
	Stop() error
}
