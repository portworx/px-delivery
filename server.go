package main

import (
	"fmt"
	"log"
	"os"

	//local packages
	"porxbbq.com/lib"

	//Mongo

	"go.mongodb.org/mongo-driver/mongo"

	//html templates

	"net/http"

	"github.com/gorilla/sessions"

	//bcrypt hash and salt

	//prometheus
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("kefue-secret-198")
	store = sessions.NewCookieStore(key)
	//mongodb client declaration
	client      *mongo.Client
	clientError error
	certString  string = ""
	mongoHost   string = os.Getenv("MONGO_HOST")
	mongoUser   string = os.Getenv("MONGO_USER")
	mongoPass   string = os.Getenv("MONGO_PASS")
	mongoTLS    string = os.Getenv("MONGO_TLS")
	kafkaHost   string = os.Getenv("KAFKA_HOST")
	kafkaPort   string = os.Getenv("KAFKA_PORT")
	mysqlHost   string = os.Getenv("MYSQL_HOST")
	mysqlUser   string = os.Getenv("MYSQL_USER")
	mysqlPass   string = os.Getenv("MYSQL_PASS")
	mysqlPort   string = os.Getenv("MYSQL_PORT")
)

func main() {
	// check DB Connections on startup
	lib.MongoCheck(mongoHost, mongoUser, mongoPass, mongoTLS)
	//lib.KafkaCheck(kafkaHost, kafkaPort)
	lib.MysqlCheck(mysqlHost, mysqlUser, mysqlPass, mysqlPort)

	//web
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	//http.HandleFunc("/", lib.TestHandler)
	http.HandleFunc("/loyalty", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path += ".html"
		fileServer.ServeHTTP(w, r)
	})
	http.HandleFunc("/healthz", lib.HealthHandler)    //used
	http.Handle("/metrics", promhttp.Handler())       //used
	http.HandleFunc("/login", lib.LoginHandler)       //used
	http.HandleFunc("/logout", lib.LogoutHandler)     //used
	http.HandleFunc("/register", lib.RegisterHandler) //used
	http.HandleFunc("/redirect_login", lib.OrderLoginHandler)
	http.HandleFunc("/order_history", lib.MyOrderHandler)              //used
	http.HandleFunc("/test", lib.TestHandler)                          //used for testing
	http.HandleFunc("/contact", lib.ContactHandler)                    //used
	http.HandleFunc("/order", lib.OrderHandler)                        //general order page where you select restaurants
	http.HandleFunc("/pxbbq_order", lib.PxbbqOrderHandler)             //Store Ordering
	http.HandleFunc("/mcd_order", lib.McdOrderHandler)                 //Store Ordering
	http.HandleFunc("/centralperk_order", lib.CentralperkOrderHandler) //Store Ordering

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
