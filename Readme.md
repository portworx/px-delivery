# Portworx Delivery (Food Delivery) Demo App

## Purpose:
The Portworx Delivery App is useful for demonstrating Portworx Data Services where you need many different data services for your application to run correctly. 

This application consists of a web application running golang. The Authentication system for the food ordering site, uses a MongoDB database. The ordering system submits orders to a Kafka queue where presumably many consumers would be reading this data. One of the consumers also written in golang will pull this data from kafka and push it into a mysql database. The golang web application performs reads from the mysql database. This pattern makes writes eventually consistent.

![PX Delivery Architecture Diagram](./static/px-delivery-arch.png)

# Instructions
## Instructions: Kubernetes

This application requires data services to be deployed prior to deploying the application. Before deploying this application, deploy a MongoDB Database, a Kafka Queue, and a Mysql Database. These databases can all be deployed by Portworx Data Services or manually. 

Once the data services have been deployed, you can use the Kubernetes manifetsts in the [Kubernetes](./kubernetes/) folder to deploy the applications.

Update the Kubernetes manifests with connection information for the `Kafka` instance, `MongoDB` database, and `MySQL` database. Specifically these environment variables:

`MONGO_HOST`<br/>
`MONGO_USER`<br/>
`MONGO_PASS`<br/>
`KAFKA_HOST`<br/>
`KAFKA_PORT`<br/>
`MYSQL_HOST`<br/>
`MYSQL_USER`<br/>
`MYSQL_PASS`<br/>
`MYSQL_PORT`<br/>



Then run the manifests in numbered order. 

This will deploy:

- A namespace 
- Portworx Delivery App (From Dockerhub) which uses MongoDB for auth and pushes to Kafka for Orders.
- Kafka Consumer that pulls from Kafka and pushes to MySQL

## Instructions: Docker

To run the Portworx Delivery App in docker run:

``` bash
docker run -e "KAFKA_HOST=172.19.0.3" -e "KAFKA_PORT=9092" -e "MYSQL_HOST=127.0.0.1" -e "MYSQL_USER=root", -e "MYSQL_PASS=porxie", -e "MYSQL_PORT=3306" eshanks16/pxdelivery:v1
```

## Instructions Golang

1. Clone the repository
2. `cd` to the px-delivery directory
3. Run
   ``` bash
    go mod tidy
    go run server.go
   ```

![Porx Image](./static/assets/img/stork.png)
