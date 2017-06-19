package lobby

// Plugin is a generic lobby plugin.
type Plugin interface {
	// Unique name of the plugin
	Name() string

	// Gracefully closes the plugin
	Close() error

	// Wait for the plugin to quit or crash. This is a blocking operation.
	Wait() error
}
