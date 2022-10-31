# Portworx BBQ Demo App

## Purpose:
The Portworx BBQ app is a Golang web application with a Mongo Database Backend. The files in this repository include the application and dependencies, along with a Dockerfile to build the container, and Kubernetes manifests to deploy in a cluster.

# Instructions
## Instructions: Kubernetes

To run in a Kubernetes cluster just `cd` to the [Kubernetes](./kubernetes/) folder and run the manifests in numbered order. The defaults should work if manifests are unchanged. If you wish to change the connection strings for the Portworx BBQ web app connection to MongoDB, change the ENV variables in the manifest. `(MONGO_HOST, MONGO_USER,MONGO_PASS)`

This will deploy:

- A namespace
- Mongo DB Container (from dockerhub)
- Portworx BBQ App (from dockerhub)

## Instructions: Docker

To run the Portworx BBQ App in docker run:

``` bash
docker run --name pxbbq -e MONGO_HOST=localhost -e MONGO_USER=porxie -e MONGO_PASS=porxie eshanks16/pxbbq:v1
```

## Instructions Golang

1. Clone the repository
2. `cd` to the porxbbq directory
3. Run
   ``` bash
    go mod tidy
    go run server.go
   ```

![Porx Image](./static/images/porx.png)