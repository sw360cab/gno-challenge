# K8S Cluster (Experimental)

A potential setup for a Kubernetes cluster.

## Prerequisites

* K8S (+ kubectl)
* [Kind](https://kind.sigs.k8s.io/)

## Using Kind

In order to test a minimal setup of a K8s cluster in a local enironment, it may be useful to use [Kind](https://kind.sigs.k8s.io/)

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

* Generate Config Map with static files

  ```bash
  kubectl apply -k config/configmap
  ```

* Generate Volumes

  ```bash
  kubectl apply -f config/volumes.yaml
  ```

## Running the cluster

* Spin up all the services

  ```bash
  kubectl create -f deploys/
  ```

* Run Stress Tests

  ```bash
  kubectl create -f jobs/supernova.yaml
  ```

* Expose dashboard service manually

  ```bash
  kubectl port-forward service/grafana 3000:3000
  ```

* Check out the Grafana dashboard by visiting [http://127.0.0.1:3000/dashboards](http://127.0.0.1:3000/dashboards) and after logging in navigate to the `Gnoland Dashboard`
(use the password defined into `grafana.ini` file for the `admin` user)