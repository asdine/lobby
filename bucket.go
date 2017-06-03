package lobby

// Errors.
const (
	ErrBackendNotFound     = Error("backend not found")
	ErrKeyNotFound         = Error("key not found")
	ErrBucketNotFound      = Error("bucket not found")
	ErrBucketAlreadyExists = Error("bucket already exists")
)

// An Item is a key value pair saved in a bucket.
type Item struct {
	Key   string
	Value []byte
}

// A Bucket manages a collection of items.
type Bucket interface {
	// Put a key value pair in the bucket. It returns the created or updated item.
	Put(key string, value []byte) (*Item, error)
	// Get an item from the bucket.
	Get(key string) (*Item, error)
	// Delete an item from the bucket.
	Delete(key string) error
	// Get the paginated list of items.
	Page(page int, perPage int) ([]Item, error)
	// Close the bucket. Can be used to close sessions if required.
	Close() error
}

// A Backend is able to create buckets that can be used to store and fetch data.
type Backend interface {
	// Get a bucket by name.
	Bucket(name string) (Bucket, error)
	// Close the backend connection.
	Close() error
}

// A Registry manages the buckets, their configuration and their associated Backend.
type Registry interface {
	Backend

	// Register a backend under the given name.
	RegisterBackend(name string, backend Backend)
	// Create a bucket and register it to the Registry.
	Create(backendName, bucketName string) error
}
