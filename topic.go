package lobby

// Errors.
const (
	ErrBackendNotFound    = Error("backend not found")
	ErrTopicNotFound      = Error("topic not found")
	ErrTopicAlreadyExists = Error("topic already exists")
)

// A Message is a key value pair saved in a topic.
type Message struct {
	Group string
	Value []byte
}

// A Topic manages a collection of items.
type Topic interface {
	// Send a message in the topic.
	Send(*Message) error
	// Close the topic. Can be used to close sessions if required.
	Close() error
}

// TopicFunc creates a topic from a send function.
func TopicFunc(fn func(*Message) error) Topic {
	return &topicFunc{fn}
}

type topicFunc struct {
	fn func(*Message) error
}

func (t *topicFunc) Send(m *Message) error {
	return t.fn(m)
}

func (t *topicFunc) Close() error {
	return nil
}

// A Backend is able to create topics that can be used to store data.
type Backend interface {
	// Get a topic by name.
	Topic(name string) (Topic, error)
	// Close the backend connection.
	Close() error
}

// A Registry manages the topics, their configuration and their associated Backend.
type Registry interface {
	Backend

	// Register a backend under the given name.
	RegisterBackend(name string, backend Backend)
	// Create a topic and register it to the Registry.
	Create(backendName, topicName string) error
}
