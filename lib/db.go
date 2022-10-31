package lib

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client      *mongo.Client
	certString  string = ""
	clientError error
)

func getMongoClient(mongoHost string, mongoUser string, mongoPass string, mongoTLS string) (*mongo.Client, error) {

	//mongoTLS string is required on DocumentDB. If running against DocumentDB ensure that the MONGO_TLS enviornment variable is not an empty string!
	if mongoTLS != "" {
		certString = "/?ssl=true&ssl_ca_certs=rds-combined-ca-bundle.pem&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false"
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + mongoUser + ":" + mongoPass + "@" + mongoHost + ":27017" + certString)
	fmt.Println("Connection String is: " + "mongodb://" + mongoUser + ":" + mongoPass + "@" + mongoHost + ":27017" + certString)

	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	return client, clientError
}

func MongoCheck(mongoHost string, mongoUser string, mongoPass string, mongoTLS string) {
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
}
