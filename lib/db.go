package lib

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client        *mongo.Client
	certString    string = ""
	clientError   error
	mongoInitUser string = os.Getenv(("MONGO_INIT_USER"))
	mongoInitPass string = os.Getenv(("MONGO_INIT_PASS"))
)

func getMongoClient(mongoHost string, mongoUser string, mongoPass string, mongoTLS string) (*mongo.Client, error) {

	//mongoTLS string is required on DocumentDB. If running against DocumentDB ensure that the MONGO_TLS enviornment variable is not an empty string!
	if mongoTLS != "" {
		certString = "/?ssl=true&ssl_ca_certs=rds-combined-ca-bundle.pem&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false"
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + mongoUser + ":" + mongoPass + "@" + mongoHost + ":27017" + certString)
	fmt.Println("Connection String is: " + "mongodb://" + mongoUser + ":" + mongoPass + "@" + mongoHost + ":27017" + certString)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	fmt.Println(nil)
	if err != nil {
		fmt.Println(("No DB found, Initializing MongoDB"))
		//InitMongoDB(mongoInitUser, mongoInitPass)
	}
	return client, clientError
}

func MongoCheck(mongoHost string, mongoUser string, mongoPass string, mongoTLS string) {
	fmt.Println("Started MongoCheck")
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		InitMongoDB(mongoInitUser, mongoInitPass, mongoUser, mongoPass)
		//log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
}

func MysqlCheck(mysqlHost string, mysqlUser string, mysqlPass string, mysqlPort string) {
	dsn := mysqlUser + ":" + mysqlPass + "@tcp(" + mysqlHost + ":" + mysqlPort + ")/delivery"
	_, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MySQL!")
	}
}

func disconnect(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func createMongoUser(client *mongo.Client, user, password string) {
	log.Printf("Creating user %s.", user)
	createUserCmd := bson.D{
		{"createUser", user},
		{"pwd", password},
		{"roles", bson.A{
			bson.D{{"role", "dbAdminAnyDatabase"}, {"db", "admin"}},
			bson.D{{"role", "readWriteAnyDatabase"}, {"db", "admin"}},
		}},
	}
	if err := client.Database("admin").RunCommand(context.TODO(), createUserCmd).Err(); err != nil {
		fmt.Println(err)
	}
}

func createMongoDatabase(client *mongo.Client, dbName string) *mongo.Database {
	log.Printf("Creating database %s.", dbName)

	return client.Database(dbName)
}

func InitMongoDB(mongoInitUser string, mongoInitPass string, mongoUser string, mongoPass string) {
	// By default, the PDS user only has permission to create other users, not read/write to any databases.
	// So must create a new user for this loadtest.
	client, err := getMongoClient(mongoHost, mongoInitUser, mongoInitPass, mongoTLS)
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { disconnect(client) }()

	fmt.Println("### Calling CreateMongoUser ###")
	fmt.Println("MONGO USER IS : " + mongoUser)
	fmt.Println("MONGO USER IS : " + mongoPass)
	createMongoUser(client, mongoUser, mongoPass)

	// Connect again with the mongoUser user, so we can read and write.
	pxclient, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	err = pxclient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { disconnect(pxclient) }()

	//create database for px-delivery
	dbName := "pxdelivery"
	fmt.Println("### Calling CreateMongoDatabase ###")
	createMongoDatabase(pxclient, dbName)
}
