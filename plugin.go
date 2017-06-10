package lobby

// Plugin is a generic lobby plugin.
type Plugin interface {
	Name() string
	Close() error
}
