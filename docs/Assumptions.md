# Assuptions & Limitations

## Design decisions

* it is covered a single _chain id_ at the moment
* the current implementation is focusing on processing transaction metrics from block 1
* it is supposed having a single instance of `tx-indexer` service
* multiple instances of `gnoland` service are subject to how `tx-indexer` service is able to handle them
* there is a single instance of houston, which is referring to the `tx-indexer` as single source of truth
* docker compose services are designed to employ `entrypoint` directive, so that the arguments of each service can be easily overridden (see [custom configuration](../README.md#custom-configuration))

## Gnoland

* It was not found a way via the command line to overide the `listen address` of the RPC server.
The value is set in the source code at `tm2/pkg/bft/rpc/config/config.go` and it can be overridden using the key `laddr` in the TOML file.
However it appears not possible to create a TOML file with only this single entry, and it is complicated to understand the correct values for all the other requested config entries.

  A suggested workaround to generate a valid TOML file was to start Gnoland with the argument `--skip-start`.
A sample of TOML file is provided in [gnoland/config.toml](../gnoland/config.toml) and can be modified and used with the current docker image
with a [custom configuration](../README.md#custom-configuration)

## Houston

* at the moment it is supposed that there is only a single message in each transaction

* only aggregated data are stored, any detail on each transaction is not saved

* aggragated metrics are saved in-memory. Although a huge amount of transaction should be expected, transactions details are not saved, so aggregated metrics, even if very large, will be magnitude of order lower in memory footprint respect to transaction details

* it is losely coupled with the `tx-indexer`: no information on the consistency of data from the `tx-indexer` is available.

* it is ephemeral, any data is not persisting on restart. This is because the `tx-indexer` which is actually persisting transaction data and should be considered, with the current architecture, a single source of truth for the application itself. The limitation is that in case of failures of the application, data are lost and should be processed back from block 1 from the `tx-indexer`, but on the other side they remain perfectly consistent

* persisting aggregated data in memory may cause inconsistencies whether the transaction indexer crashes. That is exactly because aggregated metrics does not store any kind of details about transactions. In the future it may be useful to thing carefully how to address this issue

* the GraphQL client is a POC itself, it should be finely configured to handle reconnection and retries, especially for the subscription part.
Currently employed GraphQL client library has a configurable retry timeout as long as other connection parameters

* data are exposed as a key-value JSON array from the HTTP endpoints. This was made to deal with how Grafana expects to receive data to be shown in the dashboard. the conversion from hasp map to ordered JSON array for each request is expemsive and should be improved

* ordering results each time a request is made may not scale well and the task can be left to the UI itself (check `senders` panel in Grafana)
or either data can be inserted in an custom implementation / library for ordered hash maps

### Future implementations

* storing metrics in-memory does not scale well, a decision should be taken in the direction of avoid bloating memory.
On the other hand it should be avoided to introduce a new level of persistence beyond the `tx-indexer`.
Furtherly to deal with data consistency, either references to single blocks / transactions should be mainteined, making
the application tightly coupled with the `tx-indexer`, either a mechanism that is able to understand if `tx-indexer` is providing
new data or reconstructing from scratch should be found.

* to achieve scalability the GraphQL client and the HTTP server can be decoupled and a layer of (internal) persistance can be added.
This decision should be picked with care to avoid over engineering the architecture of the service itself

* in order to add more metrics it is enough to add more methods to the interface, that will be implemented then, and to expose the relative HTTP endpoints

* data structure employed to model aggregated metrics should be carefully chosen. Hash maps are suitable in a demo-like scenario,
but may become unfit when dealing for example with multiple chains of blocks. Even transformations at the UI level may be considered

* initial query of pre-existing data should be scaled by paginating somehow the requests to the `tx-indexer`
