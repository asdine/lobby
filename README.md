# Lobby

[![GoDoc](https://godoc.org/github.com/asdine/lobby?status.svg)](https://godoc.org/github.com/asdine/lobby)
[![Go Report Card](https://goreportcard.com/badge/github.com/asdine/lobby)](https://goreportcard.com/report/github.com/asdine/lobby)

Lobby is an open-source pluggable platform for data delivery.

## Overview

At the core, Lobby is a framework to assemble network APIs and backends.
It provides several key features:

- **Key-Value store**: Applications can create buckets they can use to save or fetch data using the API of their choice. A bucket is bound to a particular backend, Lobby can route data to the right backend so applications can target multiple backends using different buckets.
- **Protocol and backend agnostic**: Data can be received using HTTP, gRPC, asynchronous consumers or by any other means and delivered to any database or proxied to another service. Lobby provides a solid framework that links everything together.
- **Plugin based architecture**: Lobby can be extended using plugins. New APIs and backends can be written using any language that supports gRPC.

## How it works

### Bucket

Lobby uses a concept of **Bucket** to store and fetch data. Each bucket is associated with a specific backend and provide an unified API to read, write and delete values.

### Backend

A backend is the storage unit used by Lobby. It usually represents a datastore but can litteraly be anything that satisfies the backend interface, like an http proxy, a file or a broker.
By default, Lobby is shipped with a builtin BoltDB backend and provides MongoDB and Redis backends as plugins.

### Entrypoints

Lobby can run multiple servers at the same time, each providing a different entrypoint to manipulate buckets. Those entrypoints can create and manipulate all or part of Lobby's buckets.
By default, Lobby runs a gRPC server which is the main communication system, also used to communicate with plugins. The HTTP and NSQ are provided as plugins.

## Usage

Running Lobby:

```sh
lobby run
```

The previous command runs the gRPC server and BoltDB backend. This is not really useful, Lobby is much more powerful with plugins.

To run Lobby with plugins:

```sh
lobby run --server=http --server=nsq --backend=mongo --backend=redis
```

The previous command adds an HTTP server, an NSQ consumer, a MongoDB and a Redis backend.

```
+------+                    +-----------+
| HTTP +-+                +-+  MONGODB  |
+------+ |                | +-----------+
         |                |
         |   +---------+  |
         +---+         +--+
+------+     |         |    +-----------+
| gRPC +-----+  LOBBY  +----+   REDIS   |
+------+     |         |    +-----------+
         +---+         +--+
         |   +---------+  |
         |                |
+------+ |                | +-----------+
| NSQ  +-+                +-+  BOLTDB   |
+------+                    +-----------+
```

Currently, Lobby contains no bucket.

The following command creates a bucket with a MongoDB backend using the HTTP API:

```sh
curl -X POST -d '{"name": "colors"}' http://localhost:5657/v1/buckets/mongo
```

Once the bucket is created, data can sent and fetched.

The following command will put the key `blue` in the `colors` bucket. Data can be of any type, Lobby will always turn it into valid JSON if it's not already the case.

```sh
curl -X PUT -d 'There is no blue without yellow and without orange.' http://localhost:5657/v1/b/colors/blue
```

Getting a key will always output valid JSON:

```sh
$ curl http://localhost:5657/v1/b/colors/blue
"There is no blue without yellow and without orange."
```
