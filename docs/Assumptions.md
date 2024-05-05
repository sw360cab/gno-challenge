# Assumptions & Limitations

## Design decisions

* it is covered a single _chain id_ at the moment
* only transaction messages routed to `vm` are considered to make aggregated metrics easier to verify, however including `bank` routed messages is trivial
* the current implementation is focusing on processing transaction metrics from block 1
* it is supposed having a single instance of `tx-indexer` service
* multiple instances of `gnoland` service are subject to how `tx-indexer` service is able to handle them
* there is a single instance of houston, which is referring to the `tx-indexer` as single source of truth
* docker compose services are designed to employ `entrypoint` directive, so that the arguments of each service can be easily added, changed or removed (see [custom configuration](../README.md#custom-configuration))

## Gnoland

* It was not found a way via the command line to override the `listen address` of the RPC server.
The value is set in the source code at `tm2/pkg/bft/rpc/config/config.go` and it can be overridden using the key `laddr` in the TOML file.
However it appears not possible to create a TOML file with only this single entry, and it is complicated to understand the correct values for all the other requested config entries.

  A suggested workaround to generate a valid TOML file was to start Gnoland with the argument `--skip-start`.
A sample of TOML file is provided in [gnoland/config.toml](../gnoland/config.toml) and can be modified and used with the current docker image
with a [custom configuration](../README.md#custom-configuration)

## Houston

* at the moment it is supposed that there is only a single message in each transaction

* only aggregated data are stored, any detail on each transaction is not saved

* aggregated metrics are saved in-memory. Although a huge amount of transactions should be expected, their details are not saved, so aggregated metrics, even if very large, will be magnitude of order lower in memory footprint respect to transaction details

* it is loosely coupled with the `tx-indexer`: no information on the consistency of data from the `tx-indexer` is available.

* it is ephemeral, any data is not persisting on restart. This is because the `tx-indexer` which is actually persisting transaction data should be considered, with the current architecture, a single source of truth for the application itself. The limitation is that in case of failures of the application, data is lost and should be processed back from block 1 from the `tx-indexer`.
If `houston` persists metrics in case of fault of the `tx-indexer`, there should be a mechanism that reset the metrics and get them back while the `tx-indexer` rebuilds itself. Whereas in case the `houston` application crashes it will not be possible to verify consistency or data loss if not persisting a reference to the transactions, thus somehow duplicating data already stored in the `tx-indexer`.
In the future it may be useful to analyze carefully how to address this issue

* the GraphQL client is a POC itself, it should be finely configured to handle reconnection and retries, especially for the subscription part.
Currently employed GraphQL client library has a configurable retry timeout as long as other connection parameters

* data are exposed as a key-value JSON array from the HTTP endpoints. This was made to deal with how Grafana expects to receive data to be shown in the dashboard. the conversion from hasp map to ordered JSON array for each request is expensive and should be improved

* ordering results each time a request is made may not scale well and the task can be left to the UI itself (check `senders` panel in Grafana)
or either data can be inserted in an custom implementation / library for ordered hash maps

### Future implementations

* storing metrics in-memory does not scale well, a decision should be taken in the direction of avoiding bloating memory.
On the other hand it should be avoided to introduce a new level of persistence beyond the `tx-indexer`.
Furtherly to deal with data consistency, either references to single blocks / transactions should be maintained, making
the application tightly coupled with the `tx-indexer`, either a mechanism that is able to understand if `tx-indexer` is providing
new data or reconstructing from scratch should be found.

* to achieve scalability the GraphQL client and the HTTP server can be decoupled and a layer of (internal) persistence can be added.
This decision should be picked with care to avoid over engineering the architecture of the service itself

* in order to add more metrics it is enough to add more methods (to be implemented then) to the interface and to expose the relative HTTP endpoints. Then the given aggregate metric may require to update the subscription callback itself, otherwise methods proxying direct queries to GraphQL server can be crafted

* data structure employed to model aggregated metrics should be carefully chosen. Hash maps are suitable in a demo-like scenario,
but may become unfit when dealing for example with multiple chains of blocks. Even transformations at the UI level may be considered

* initial query of pre-existing data should be scaled by paginating somehow the requests to the `tx-indexer`

* either the JSON RPC service can be considered to subscribe or query data from the `tx-indexer`

* even if Grafana is showing fancy and elegant dashboards, it is still very complex with plenty of options and settings. It is not trivial to configure UIs in a dashboard. Other alternatives may be considered to evaluate other solutions having less futures and options to configure
