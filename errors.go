package lobby

// Error represents a Lobby error.
type Error string

// Error returns the error message.
func (e Error) Error() string {
	return string(e)
}
