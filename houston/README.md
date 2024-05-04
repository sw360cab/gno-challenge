# Houston - We have a problem

The purpose of `houston` is to aggregate data fetched by the Gno `transaction indexer`
and to bridge between:

* the GraphQL interface of the `transaction indexer`
* the `Grafana` dashboard that visualizes data

There are two main items of the application:

* the GraphQL client that `subscribes` to the GraphQL server of the `tx_indexer`
* a `Gin` router to expose aggregated data via HTTP endpoints to the dashboard service.

Concurrency is avoided by protecting the `critical section` with exclusive access (Mutex).
This section is the handler of the responses by the GraphQL server, it is in charge of updating internal data structures whenever a new transaction is received in the `transaction indexer`.

## GraphQL

The application deals with the GraphQL server at two different levels:

* an event driven level, which intercepts transactions received on the `transaction indexer` in (near) real-time
* a static query level, which is in charge of gathering pre-existing data in case the application may crash

## Assumptions & Limitations

Check [docs/Assumptions.md](../docs/Assumptions.md#houston)
