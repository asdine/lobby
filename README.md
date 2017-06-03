# Lobby

Lobby is an open-source pluggable platform for data delivery.

## Overview

At the core, Lobby is a framework to assemble network APIs and backends.
It provides several key features:

- **Key-Value store**: Applications can create buckets they can use to save or fetch data using the API of their choice. A bucket is bound to a particular backend, Lobby can route data to the right backend so applications can target multiple backends using different buckets.
- **Protocol and backend agnostic**: Data can be received using HTTP, gRPC, asynchronous consumers or by any other means and delivered to any database or proxied to another service. Lobby provides a solid framework that links everything together.
- **Plugin based architecture**: Lobby can be extended using plugins. New APIs and backends can be written using the Go plugin system or any language that supports gRPC. By default, Lobby provides gRPC and HTTP APIs and a BoltDB backend.
