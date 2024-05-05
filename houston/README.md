# Houston - We have a problem

Houston is an ephemeral real-time and event driven metrics aggregator.

The purpose of `houston` is to aggregate transaction data fetched by the Gno `transaction indexer` starting from block 1 and to bridge between:

* the GraphQL interface of the `transaction indexer`
* the `Grafana` dashboard that visualizes data in a browser.

There are two main items of the application:

* the GraphQL client that `subscribes` to the GraphQL server of the `tx_indexer`
* a `Gin` router to expose aggregated data via HTTP endpoints to the dashboard service.

Due to the potential large amount of data that can be received, concurrency issues are faced by protecting the `critical section` with exclusive access (Mutex).
This section is the event handler of the responses by the GraphQL server, it is in charge of updating internal data structures upon receiving events from the GraphQL server of the `transaction indexer`, which triggers events whenever a new transaction has been processed.

## GraphQL

The application deals with the GraphQL server at two different levels:

* an event driven level, which intercepts transactions received on the `transaction indexer` in (near) real-time
* a static query level, which is in charge of gathering pre-existing data in case the application may crash

## Build

Import the required libraries via `go mod`

```bash
go mod download
```

## Running

Bootstrap the services of `houston` via:

```bash
go run main.go
```

## Testing

A minimal and trivial test suite is provided.

```bash
go test -v ./...
```

## Assumptions & Limitations

Check out [Assumptions](../docs/Assumptions.md#houston)
