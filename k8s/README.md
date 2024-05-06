# K8S Cluster (Experimental)

A potential setup for a Kubernetes cluster.

## Prerequisites

* K8S (+ kubectl)
* [Kind](https://kind.sigs.k8s.io/)

## Using Kind

In order to test a minimal setup of a K8s cluster in a local environment, it may be useful to use [Kind](https://kind.sigs.k8s.io/)

* Build the Cluster with Kind

  ```bash
  kind create cluster --name gnolive --config=kind/kind.yaml
  ```

* Get cluster info

  ```bash
  kubectl cluster-info --context kind-gnolive
  ```

* (opt.) Delete cluster using

  ```bash
  kind delete cluster --name gnolive
  ```

## Configuration

* create the file `secrets/grafana.ini` containing only a plain text password for the Grafana dashboard
(just a plain string no key/value, no quotes)

* Generate secrets

  ```bash
  kubectl apply -k secrets/
  ```

* Generate Config Map for static config files

  ```bash
  kubectl apply -k configmaps/
  ```

* Generate Volumes

  ```bash
  kubectl apply -f storage/
  ```

## Running the cluster

* Spin up all the services

  ```bash
  kubectl apply -f deploys/
  ```

* Run Stress Tests

  ```bash
  kubectl apply -f jobs/supernova.yaml
  ```

* Check out the Grafana dashboard by visiting [http://127.0.0.1:30000/dashboards](http://127.0.0.1:30000/dashboards) and after logging in navigate to the `Gnoland Dashboard`
(use the password defined into `grafana.ini` file for the `admin` user)

Note: Cluster created using `Kind` is configured to expose port 30000 on the control plane host using `extraPortMappings`.
The port is deliberately picked to work _out of the box_ on any host, but it is influenced by a combination of the versions of Docker and Kubernetes.

* (alternatively) Expose dashboard service manually and access dashboard at port 3000

  ```bash
  kubectl port-forward service/grafana 3000:3000
  ```
