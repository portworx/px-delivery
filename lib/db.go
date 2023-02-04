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
	//"github.com/confluentinc/confluent-kafka-go/kafka"
	//"github.com/segmentio/kafka-go"
)

var (
	client        *mongo.Client
	certString    string = ""
	clientError   error
	mongoUser     string = "porxie"
	mongoPass     string = "porxie"
	mongoDBName   string = "delivery"
	mongoInitUser string = os.Getenv(("MONGO_INIT_USER"))
	mongoInitPass string = os.Getenv(("MONGO_INIT_PASS"))
	kafkaHost     string = os.Getenv("KAFKA_HOST")
	kafkaPort     string = os.Getenv("KAFKA_PORT")
	kafkaUser     string = os.Getenv("KAFKA_USER")
	kafkaPass     string = os.Getenv("KAFKA_PASS")
)

type PxOrder struct {
	OrderId     int    `bson:"orderid,omitempty"`
	Email       string `bson:"email,omitempty"`
	Main        string `bson:"main,omitempty"`
	Side1       string `bson:"side1,omitempty"`
	Side2       string `bson:"side2,omitempty"`
	Drink       string `bson:"drink,omitempty"`
	Restaurant  string `bson:"restaurant,omitempty"`
	Date        string `bson:"date,omitempty"`
	Street1     string `bson:"street1,omitempty"`
	Street2     string `bson:"street2,omitempty"`
	City        string `bson:"city,omitempty"`
	State       string `bson:"state,omitempty"`
	Zip         string `bson:"zip,omitempty"`
	OrderStatus string `bson:"orderstatus,omitempty"`
}

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

func DbCheck() {
	//Check to see if MongoDB database is ready, if not create one.
	fmt.Println("Checking Mongo Database")
	mongoCheck, err := mongoCheck(mongoHost, mongoInitUser, mongoInitPass, mongoTLS)
	if err != nil {
		fmt.Println("The Mongo Host was unreachable.")
		log.Fatal(err)
	} else if !mongoCheck {
		fmt.Println("Mongo Host Reachable")
		fmt.Println("Initializing Database : " + mongoDBName)
		createMongoUser(mongoHost, mongoInitUser, mongoInitPass, mongoUser, mongoPass)
	}

	//check to see if MySQL is ready
	fmt.Println("Checking MySQL Database")
	mysqlCheck(mysqlHost, mysqlUser, mysqlPass, mysqlPort)

	//Initialize Kafka
	KafkaInit()

	//create KAFKA Topic
	//brokers := kafkaHost + ":" + kafkaPort
	//if err := createTopic(brokers, "order", kafkaUser, kafkaPass); err != nil {
	//	fmt.Printf("Failed to create topic: %v\n", err)
	//}
}

func mongoCheck(mongoHost string, mongoInitUser string, mongoInitPass string, mongoTLS string) (bool, error) {
	fmt.Println("Started MongoCheck")
	client, err := getMongoClient(mongoHost, mongoInitUser, mongoInitPass, mongoTLS)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return false, err
	}
	fmt.Println("Connected to MongoDB!")

	dbList, err := client.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		return false, err
	}

	for _, dbName := range dbList {
		if dbName == mongoDBName {
			return true, nil
		}
	}

	return false, nil
}

func mysqlCheck(mysqlHost string, mysqlUser string, mysqlPass string, mysqlPort string) {
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

func createMongoUser(mongoHost string, mongoInitUser string, mongoInitPass string, mongoUser string, mongoPass string) {
	log.Printf("Creating user %s.", mongoUser)
	client, err := getMongoClient(mongoHost, mongoInitUser, mongoInitPass, mongoTLS)
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("COULD NOT CREATE MONGO USER " + mongoUser)
		return
	}
	defer func() { disconnect(client) }()

	createUserCmd := bson.D{
		{"createUser", mongoUser},
		{"pwd", mongoPass},
		{"roles", bson.A{
			bson.D{{"role", "dbAdminAnyDatabase"}, {"db", "admin"}},
			bson.D{{"role", "readWriteAnyDatabase"}, {"db", "admin"}},
		}},
	}
	if err := client.Database("admin").RunCommand(context.TODO(), createUserCmd).Err(); err != nil {
		fmt.Println(err)
	}
}

//func createTopic(brokers, topic, username, password string) error {
//	// Create a new Kafka config object
//	config := &kafka.ConfigMap{
//		"bootstrap.servers": brokers,
//		"security.protocol": "sasl_plaintext",
//		"sasl.mechanisms":   "PLAIN",
//		"sasl.username":     kafkaUser,
//		"sasl.password":     kafkaPass,
//	}
//
//	fmt.Println("brokers are : " + brokers)
//	// Create a new Kafka admin client
//	admin, err := kafka.NewAdminClient(config)
//	if err != nil {
//		fmt.Println(err)
//		return fmt.Errorf("failed to create admin client: %v", err)
//	}
//	defer admin.Close()
//
//	fmt.Println("Topic name is : " + topic)
//
//	// Create the "order" topic if it doesn't exist
//	results, err := admin.CreateTopics(
//		context.TODO(),
//		[]kafka.TopicSpecification{{
//			Topic:             topic,
//			NumPartitions:     1,
//			ReplicationFactor: 1,
//			Config:            map[string]string{},
//		}},
//		kafka.SetAdminOperationTimeout(5*1000),
//	)
//	if err != nil {
//		return fmt.Errorf("failed to create topics: %v", err)
//	}
//
//	if results[0].Error.Code() != kafka.ErrNoError {
//		return fmt.Errorf("failed to create topic: %v", results[0].Error)
//	}
//
//	fmt.Println("Kafka Topic Created")
//	return nil
//}
