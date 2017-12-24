package etcd

import (
	"context"
	"log"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/asdine/lobby"
	"github.com/asdine/lobby/etcd/internal"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

var _ lobby.Registry = new(Registry)

// NewRegistry returns an etcd Registry.
func NewRegistry(client *clientv3.Client, namespace string) (*Registry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespace = path.Join(strings.TrimLeft(namespace, "/"), "/")
	topicsPrefix := path.Join(namespace, "topics") + "/"
	reg := Registry{
		client:       client,
		namespace:    namespace,
		topicsPrefix: topicsPrefix,
		backends:     make(map[string]lobby.Backend),
		topics:       make(map[string]*internal.Topic),
	}

	resp, err := client.Get(ctx, topicsPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve topics at path '%s'", topicsPrefix)
	}

	for _, kv := range resp.Kvs {
		err := reg.storeTopic(kv.Key, kv.Value)
		if err != nil {
			return nil, err
		}
	}

	reg.topicsWatcher = clientv3.NewWatcher(client)
	wch := reg.topicsWatcher.Watch(context.Background(), topicsPrefix, clientv3.WithPrefix())

	reg.wg.Add(1)
	go reg.watchTopics(wch)

	return &reg, nil
}

// Registry is an etcd registry.
type Registry struct {
	client        *clientv3.Client
	namespace     string
	topicsPrefix  string
	topicsWatcher clientv3.Watcher
	topics        map[string]*internal.Topic
	mu            sync.RWMutex
	wg            sync.WaitGroup
	backends      map[string]lobby.Backend
}

// RegisterBackend registers a backend under the given name.
func (r *Registry) RegisterBackend(name string, backend lobby.Backend) {
	r.backends[name] = backend
}

func (r *Registry) watchTopics(c clientv3.WatchChan) {
	defer r.wg.Done()

	for wresp := range c {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				err := r.storeTopic(ev.Kv.Key, ev.Kv.Value)
				if err != nil {
					log.Printf("Can't decode topic %s from etcd registry\n", ev.Kv.Key)
				}
			}
		}
	}
}

func (r *Registry) storeTopic(key, value []byte) error {
	var t internal.Topic
	if err := proto.Unmarshal(value, &t); err != nil {
		return err
	}

	name := strings.TrimPrefix(string(key), r.topicsPrefix)

	r.mu.Lock()
	r.topics[name] = &t
	r.mu.Unlock()
	return nil
}

// Create a topic in the registry.
func (r *Registry) Create(backendName, topicName string) error {
	if _, ok := r.backends[backendName]; !ok {
		return lobby.ErrBackendNotFound
	}

	r.mu.Lock()
	_, ok := r.topics[topicName]
	if ok {
		r.mu.Unlock()
		return lobby.ErrTopicAlreadyExists
	}

	topic := internal.Topic{
		Name:    topicName,
		Backend: backendName,
	}

	r.topics[topicName] = &topic
	r.mu.Unlock()

	raw, err := proto.Marshal(&topic)
	if err != nil {
		return errors.Wrapf(err, "failed to encode topic %s", topicName)
	}

	_, err = r.client.Put(context.Background(), path.Join(r.topicsPrefix, topicName), string(raw))
	return errors.Wrapf(err, "failed to create topic %s", topicName)
}

// Topic returns the selected topic from the Backend.
func (r *Registry) Topic(name string) (lobby.Topic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	topic, ok := r.topics[name]
	if !ok {
		return nil, lobby.ErrTopicNotFound
	}

	backend, ok := r.backends[topic.Backend]
	if !ok {
		return nil, lobby.ErrTopicNotFound
	}

	return backend.Topic(name)
}

// Close etcd connection and registered backends.
func (r *Registry) Close() error {
	defer r.wg.Wait()

	for name, backend := range r.backends {
		err := backend.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to close backend %s", name)
		}
	}

	err := r.topicsWatcher.Close()
	if err != context.Canceled {
		return errors.Wrap(err, "failed to close etcd watcher")
	}
	return nil
}
