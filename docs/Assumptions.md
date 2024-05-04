# Assuptions & Limitations

## Design decisions

* it is covered a single chain id at the moment
* the current implementation is focusing on processing transaction metrics from block 1
* it is supposed having a single instance of `tx-indexer` service
* multiple instances of `gnoland` service are subject to how `tx-indexer` service is able to handle them
* there is a single instance of houston, which is referring to the `tx-indexer` as single source of truth
* docker compose services are designed to employ `command` directive, so that the arguments of each service can be easily overridden

## Gnoland

* It was not found a way via command line to overide the `listen address` of the RPC server.
The value is set in the source code at `tm2/pkg/bft/rpc/config/config.go` and it can be overridden using the key `laddr` in the TOML file.
However it appears not possible to create a TOML file with only this single value, and it is complicated to understand the correct values for all the other requested keys.

  A suggested workaround to generate a valid TOML file was to start Gnoland with the argument `--skip-start`.
A sample of TOML file is provided in [gnoland/config.toml](../gnoland/config.toml) and can be modified and used with the current docker image
with a [custom configuration](../README.md#custom-configuration)

## Houston

* at the moment it is supposed that there is only a single message in each transaction

* aggragated metrics are saved in-memory. Even if a huge amount of transaction should be expected, houston is not saving any transactions details, this means that even if big, aggregated metrics will be magnitude of order lower in memory footprint respect to transaction details.

* it is not persisting data on restart. This is because it is relying on `tx-indexer` which is actually persisting transaction data and which should be considered with the current architecture a single source of truth for the application itself. The limitation is that in case of failures of the application, data are lost and should be processed back from block 1 from the `tx-indexer`, but on the other side they remain perfectly consistent.

* persisting aggregated data in memory may cause inconsistency whether the transaction indexer crashes. That is exactly because aggregated metrics does not store any kind of details about transactions. In the future it may be useful to thing carefully how to address this issue.

* the GraphQL client is a POC itself, it should be finely configured to handle reconnection and retries, especially for the subscription part.
Currently employed GraphQL client library has one minute of retry timeout and the value is not configurable.

* data are exposed as a key-value JSON array from the HTTP endpoints. This was made to deal with how Grafana expects to receive data to be shown in the dashboard. the conversion from hasp map to ordered JSON array for each request is expemsive and should be improved

* ordering results each time a request is made may not scale well and the task can be left to the UI itself (check `senders` panel in Grafana)
or either data can be inserted in an custom implementation / library for ordered hash maps.

### Future implementations

* to achieve scalability of houston the GraphQL client and the HTTP server can be decoupled and a layer of (internal) persistance should be added.
This decision should be picked with care to avoid over engineering the architecture of the application itself.

* in order to add more metrics it is enough to add more methods to the interface and to expose the relative HTTP endpoint

* storing metrics in-memory does not scale well, a decision should be taken in the direction of avoid bloating memory.
On the other hand it should be avoided to introduce a new level of persistence beyond the `tx-indexer`

* data structure employed to model aggregated metrics should be carefully chose. Hash maps are suitable in a demo-like scenario,
but may become unfit when dealing for example with multiple chains of blocks.

* initial query of per-existing data should be scaled by paginating somehow the requests to `tx-indexer`
