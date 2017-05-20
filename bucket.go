package lobby

import "errors"

// Errors.
var (
	ErrKeyNotFound         = errors.New("key not found")
	ErrBucketNotFound      = errors.New("bucket not found")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
)

// An Item is a key value pair saved in a bucket.
type Item struct {
	Key  string
	Data []byte
}

// A Bucket manages a collection of items.
type Bucket interface {
	// Save a key value pair. It returns the created item.
	Save(key string, data []byte) (*Item, error)
	// Get an item from the bucket.
	Get(key string) (*Item, error)
	// Delete an item from the bucket.
	Delete(key string) error
	// Get the paginated list of items. perPage can be set to -1 to fetch all the items.
	Page(page int, perPage int) ([]Item, error)
	// Close the bucket. Can be used to close sessions if required.
	Close() error
}

// A Backend is able to create buckets that can be used to store and fetch data.
type Backend interface {
	// Get a bucket managing the given path.
	Bucket(name string) (Bucket, error)
	// Close the backend connection.
	Close() error
}

// A Registry manages the buckets, their configuration and their associated Backend.
type Registry interface {
	// Register a backend under the given name.
	RegisterBackend(name string, backend Backend)
	// Create a bucket and register it to the Registry.
	Create(backendName, bucketName string) error
	// Fetch a bucket directly from the associated Backend.
	Bucket(name string) (Bucket, error)
	// Close the registry connection.
	Close() error
}
