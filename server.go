package brazier

import "net"

// A Server serves incoming requests.
type Server interface {
	// Serve incoming requests. Must block.
	Serve(net.Listener) error

	// Stop gracefully stops the server.
	Stop() error
}
