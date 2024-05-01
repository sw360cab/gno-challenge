# Drifting through the Cosmos

Solution to GNO challenge by Sergio Matone

## Prerequisites

* Docker suite (build,compose...) >= 20.04

## Preliminary steps

* create the file `grafana/grafana.ini` containing the plain text password for Grafana dashboard

## Running

* Bootstrap the basic configuration

```bash
docker compose up -d
```

This will launch all the services, but `supernova` will fail quite immediately.
That is because we will use `supernova` as stress test and launch it multiple times.

* Launch a stress test

```bash
docker compose run --rm supernova
```

  Or

```bash
docker-compose run --rm supernova -sub-accounts 5 -transactions 500 -url http://gnoland:26657 -mode REALM_CALL
-mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
```

* Visit the Grafana dashboard at `http://127.0.0.1:3000/`
(use the credentials defined into `grafana.ini` file)

## Services of Cosmos

* `gnoland`: Gno.land blockchain node
* `gnoweb`: Gno.land web interface
* `tx-indexer`: transaction indexer
* `supernova`: network activity simulator
* `houston`: an aggragetor of transaction metrics
* `grafana`: the dashboard unit using Grafana dashboard

## Notes on Houston

* Check [Houston](houston/README.md)

## Docker Images of service

* `sw360cab/aib-gnoland`: `gnoland` image based on
([Dockerfile in git repo](https://raw.githubusercontent.com/gnolang/gno/master/Dockerfile)) using code slightly modified
* `sw360cab/aib-gnoweb`: same as `gnoland`
* `sw360cab/aib-tx-indexer`: image created using git repo [`Dockerfile`](https://raw.githubusercontent.com/gnolang/tx-indexer/main/Dockerfile)
* `sw360cab/aib-supernova`: image created using git repo source code and a custom [Dockerfile](supernova-build/supernova.Dockerfile)
* `houston`: custom multi-stage `Dockerfile` for Go applications
* `grafana`: official image provided by Grafana

## Data pipeline

This is the flow of data that is supposed to happen:

supernova -> gnoland -> tx-indexer -> houston -> dashboard

## Assumptions & Limitations

Check [Assumptions.md](Assumptions.md)
