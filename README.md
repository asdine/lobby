# Lobby

[![GoDoc](https://godoc.org/github.com/asdine/lobby?status.svg)](https://godoc.org/github.com/asdine/lobby)
[![Build Status](https://travis-ci.org/asdine/lobby.svg)](https://travis-ci.org/asdine/lobby)
[![Go Report Card](https://goreportcard.com/badge/github.com/asdine/lobby)](https://goreportcard.com/report/github.com/asdine/lobby)

Lobby is an open-source pluggable platform for data delivery.

![Credits @jrmneveu](https://user-images.githubusercontent.com/2102036/27262061-60c8c4ee-544f-11e7-96d7-17464a41fc1d.png)

## Overview

At the core, Lobby is a framework to assemble network APIs and backends.
It provides several key features:

- **Topics**: Applications can create topics they can use to save data using the API of their choice. A topic is bound to a particular backend, Lobby can route data to the right backend so applications can target multiple stores using different topics.
- **Protocol and backend agnostic**: Data can be received using HTTP, gRPC, asynchronous consumers or by any other means and delivered to any database, message broker or even proxied to another service. Lobby provides a solid framework that links everything together.
- **Plugin based architecture**: Lobby can be extended using plugins. New APIs and backends can be written using any language that supports gRPC.

## How it works

### Topic

Lobby uses a concept of **Topic** to store data. Each topic is associated to a specific backend and provide an unified API to send values.

### Backend

A backend is the storage unit used by Lobby. It usually represents a datastore or a message broker but can litteraly be anything that satisfies the backend interface, like an http proxy, a file or a memory store.
By default, Lobby is shipped with a builtin BoltDB backend and provides MongoDB and Redis backends as plugins.

### Entrypoints

Lobby can run multiple servers at the same time, each providing a different entrypoint to manipulate topics. Those entrypoints can create and manipulate all or part of Lobby's topics.
By default, Lobby runs an HTTP server and a gRPC server which is the main communication system, also used to communicate with plugins. NSQ is provided as a plugin.

## Usage

Running Lobby:

```sh
lobby run
```

The previous command runs the gRPC server, the HTTP server and the BoltDB backend.

To run Lobby with plugins:

```sh
lobby run --server=nsq --backend=mongo --backend=redis
```

The previous command adds a NSQ consumer, a MongoDB and a Redis backend.

Currently, Lobby contains no topics.

The following command creates a topic with a Redis backend using the HTTP API:

```sh
curl -X POST -d '{"name": "quotes", "backend": "redis"}' http://localhost:5657/v1/topics
```

Once the topic is created, data can be sent to it.

The following command will send the following value in the `quotes` topic.

```sh
curl -X POST -d 'There is no blue without yellow and without orange.' \
                                  http://localhost:5657/v1/topics/quotes
```
