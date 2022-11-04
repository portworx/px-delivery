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
)

func main() {
	// check DB Connection on startup
	lib.MongoCheck(mongoHost, mongoUser, mongoPass, mongoTLS)

	//web
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/loyalty", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path += ".html"
		fileServer.ServeHTTP(w, r)
	})
	http.HandleFunc("/healthz", lib.HealthHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/login", lib.LoginHandler)
	http.HandleFunc("/logout", lib.LogoutHandler)
	http.HandleFunc("/register", lib.RegisterHandler)
	http.HandleFunc("/loyaltystatus", lib.LoyaltystatusHandler)
	http.HandleFunc("/order", lib.OrderHandler)
	http.HandleFunc("/redirect_login", lib.OrderLoginHandler)
	http.HandleFunc("/myorders", lib.MyOrderHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
