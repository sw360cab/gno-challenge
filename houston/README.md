# Houston - We have a problem

The purpose of `houston` is to aggregate data fetched by the Gno `transaction indexer`
and to bridge between:

* the GraphQL interface of the `transaction indexer`
* the `Grafana` dashboard that visualizes data

There are two main items of the application:

* the GraphQL client that `subscribes` to the GraphQL server of the `tx_indexer`
* a `Gin` router to expose aggregated data via HTTP endpoints to the dashboard service.

Concurrency is avoided by protecting the `critical section` with exclusive access (Mutex).
This section is the handler of the GraphQL subscrition which receives new data and
updates internal data structures whenever a new transaction is received in the transaction indexer.
