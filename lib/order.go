package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	kafkaHost string = os.Getenv("KAFKA_HOST")
	kafkaPort string = os.Getenv("KAFKA_PORT")
)

type Order struct {
	OrderId     int    `bson:"orderid,omitempty"`
	Email       string `bson:"email,omitempty"`
	Main        string `bson:"main,omitempty"`
	Side1       string `bson:"side1,omitempty"`
	Side2       string `bson:"side2,omitempty"`
	Drink       string `bson:"drink,omitempty"`
	OrderStatus string `bson:"orderstatus,omitempty"`
}

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

type myOrderData struct {
	PageTitle string
	Orders    []Order
}

func generateOrderID() int {
	rand.Seed(time.Now().UnixNano())
	orderId := rand.Intn(100000)
	return (orderId)
}

func GetMyOrders(email string) []Order {
	fmt.Println("#############Executing Function GetMyOrders##############")
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("orders")

	filter := bson.D{{"email", email}}

	var myOrders []Order

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context.TODO(), &myOrders); err != nil {
		log.Fatal(err)
	}

	return (myOrders)

}

func registerOrder(orderNum int, email string, main string, side1 string, side2 string, drink string) {
	client, err := getMongoClient(mongoHost, mongoUser, mongoPass, mongoTLS)
	collection := client.Database("porxbbq").Collection("orders")

	//fmt.Println(email)
	entry := Order{orderNum, email, main, side1, side2, drink, "preparing"}

	insertResult, err := collection.InsertOne(context.TODO(), entry)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a Single Document: ", insertResult.InsertedID)

}

func orderStatus(w http.ResponseWriter, r *http.Request, messageData PageData) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	t, _ := template.ParseFiles("./static/order_status.html")
	if r.Method == "POST" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			t, _ = template.ParseFiles("./static/external_order_status.html")
		}
		t.Execute(w, messageData)
	}
}

func OrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["authenticated"])
	fmt.Println(session.Values["email"])
	fmt.Println(r.Method)
	//generate Order ID
	orderNum := generateOrderID()

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		fmt.Println("Not Authenticated")
	} else {
		fmt.Println("Authenticated")
	}

	if r.Method == "GET" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Order form requested, but unauthenticated; redirecting to login page.")
			t, _ := template.ParseFiles("./static/login.html")
			t.Execute(w, nil)
		} else {
			fmt.Printf("should allow order")
			t, _ := template.ParseFiles("./static/order.html")
			t.Execute(w, nil)
		}
	} else {
		r.ParseForm()
		statusData := PageData{
			PageTitle: "Order Status",
			Message:   fmt.Sprintf("Your order has been received. Order number %v", orderNum),
		}

		//write to mongo
		fmt.Printf("Order submitted by: ")
		fmt.Println(session.Values["email"].(string))
		email := session.Values["email"].(string)
		main := r.FormValue("main")
		side1 := r.FormValue("side1")
		side2 := r.FormValue("side2")
		drink := r.FormValue("drink")

		registerOrder(orderNum, email, main, side1, side2, drink)

		//Display Operation Status Page to User
		orderStatus(w, r, statusData)
	}
}

func MyOrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["email"])
	email := session.Values["email"].(string)

	myOrders := GetMyOrders(email)

	data := myOrderData{
		PageTitle: "My Order History",
		Orders:    myOrders,
	}

	t, _ := template.ParseFiles("./static/myorders.html")
	t.Execute(w, data)

}

func PxbbqOrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["authenticated"])
	fmt.Println(session.Values["email"])
	fmt.Println(r.Method)

	//generate Order ID
	orderNum := generateOrderID()

	//verify the user is authenticated and if not send them to the login page instead
	if r.Method == "GET" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Order form requested, but unauthenticated; redirecting to login page.")
			t, _ := template.ParseFiles("./static/login.html")
			t.Execute(w, nil)
		} else {
			fmt.Printf("should allow order")
			t, _ := template.ParseFiles("./static/centralperk_order.html")
			t.Execute(w, nil)
		}
	} else {
		r.ParseForm()
		statusData := PageData{
			PageTitle: "Order Status",
			Message:   fmt.Sprintf("Your order has been received. Order number %v", orderNum),
		}

		//Organize form submission data for a write to storage
		currentTime := time.Now() //used in the order submission
		email := session.Values["email"].(string)
		main := r.FormValue("main")
		side1 := r.FormValue("side1")
		side2 := r.FormValue("side2")
		drink := r.FormValue("drink")

		//retrieve address from customer's account
		myAddress := GetAddress(email)
		street1 := myAddress.Street1
		street2 := myAddress.Street2
		city := myAddress.City
		state := myAddress.State
		zipcode := myAddress.Zipcode
		orderDate := currentTime.Format("2 January 2006")

		//log order to std out - Can be used for troubleshooting
		fmt.Printf("Order submitted by: ")
		fmt.Println(session.Values["email"].(string))
		fmt.Println("Order Taken by Portworx BBQ")
		fmt.Println("Ordered : " + main)
		fmt.Println("Ordered : " + side1)
		fmt.Println("Ordered : " + side2)
		fmt.Println("Ordered : " + drink)
		fmt.Println("Street1 : " + street1)
		fmt.Println("Street2 : " + street2)
		fmt.Println("City  : " + city)
		fmt.Println("State  : " + state)
		fmt.Println("Zip  : " + zipcode)
		fmt.Print("########")

		//submit order to storage
		SubmitOrder(orderNum, orderDate, email, "PXBBQ", main, side1, side2, drink, street1, street2, city, state, zipcode)

		//Display Operation Status Page to User
		orderStatus(w, r, statusData)
	}
}

func McdOrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["authenticated"])
	fmt.Println(session.Values["email"])
	fmt.Println(r.Method)

	//generate Order ID
	orderNum := generateOrderID()

	//verify the user is authenticated and if not send them to the login page instead
	if r.Method == "GET" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Order form requested, but unauthenticated; redirecting to login page.")
			t, _ := template.ParseFiles("./static/login.html")
			t.Execute(w, nil)
		} else {
			fmt.Printf("should allow order")
			t, _ := template.ParseFiles("./static/mcd_order.html")
			t.Execute(w, nil)
		}
	} else {
		r.ParseForm()
		statusData := PageData{
			PageTitle: "Order Status",
			Message:   fmt.Sprintf("Your order has been received. Order number %v", orderNum),
		}

		//Organize form submission data for a write to storage
		currentTime := time.Now() //used in the order submission
		email := session.Values["email"].(string)
		main := r.FormValue("main")
		side1 := r.FormValue("side1")
		side2 := r.FormValue("side2")
		drink := r.FormValue("drink")

		//retrieve address from customer's account
		myAddress := GetAddress(email)
		street1 := myAddress.Street1
		street2 := myAddress.Street2
		city := myAddress.City
		state := myAddress.State
		zipcode := myAddress.Zipcode
		orderDate := currentTime.Format("2 January 2006")

		//log order to std out - Can be used for troubleshooting
		fmt.Printf("Order submitted by: ")
		fmt.Println(session.Values["email"].(string))
		fmt.Println("Order Taken by McDowells")
		fmt.Println("Ordered : " + main)
		fmt.Println("Ordered : " + side1)
		fmt.Println("Ordered : " + side2)
		fmt.Println("Ordered : " + drink)
		fmt.Println("Street1 : " + street1)
		fmt.Println("Street2 : " + street2)
		fmt.Println("City  : " + city)
		fmt.Println("State  : " + state)
		fmt.Println("Zip  : " + zipcode)
		fmt.Print("########")

		//submit order to storage
		SubmitOrder(orderNum, orderDate, email, "McDowells", main, side1, side2, drink, street1, street2, city, state, zipcode)

		//Display Operation Status Page to User
		orderStatus(w, r, statusData)
	}
}

func CentralperkOrderHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on URL:", r.URL)
	session, _ := store.Get(r, "cookie-name")
	fmt.Println(session.Values["authenticated"])
	fmt.Println(session.Values["email"])
	fmt.Println(r.Method)

	//generate Order ID
	orderNum := generateOrderID()

	//verify the user is authenticated and if not send them to the login page instead
	if r.Method == "GET" {
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			fmt.Println("Order form requested, but unauthenticated; redirecting to login page.")
			t, _ := template.ParseFiles("./static/login.html")
			t.Execute(w, nil)
		} else {
			fmt.Printf("should allow order")
			t, _ := template.ParseFiles("./static/centralperk_order.html")
			t.Execute(w, nil)
		}
	} else {
		r.ParseForm()
		statusData := PageData{
			PageTitle: "Order Status",
			Message:   fmt.Sprintf("Your order has been received. Order number %v", orderNum),
		}

		//Organize form submission data for a write to storage
		currentTime := time.Now() //used in the order submission
		email := session.Values["email"].(string)
		main := r.FormValue("main")
		side1 := r.FormValue("side1")
		side2 := r.FormValue("side2")
		drink := r.FormValue("drink")

		//retrieve address from customer's account
		myAddress := GetAddress(email)
		street1 := myAddress.Street1
		street2 := myAddress.Street2
		city := myAddress.City
		state := myAddress.State
		zipcode := myAddress.Zipcode
		orderDate := currentTime.Format("2 January 2006")

		//log order to std out - Can be used for troubleshooting
		fmt.Printf("Order submitted by: ")
		fmt.Println(session.Values["email"].(string))
		fmt.Println("Order Taken by Central Perk")
		fmt.Println("Ordered : " + main)
		fmt.Println("Ordered : " + side1)
		fmt.Println("Ordered : " + side2)
		fmt.Println("Ordered : " + drink)
		fmt.Println("Street1 : " + street1)
		fmt.Println("Street2 : " + street2)
		fmt.Println("City  : " + city)
		fmt.Println("State  : " + state)
		fmt.Println("Zip  : " + zipcode)
		fmt.Print("########")

		//submit order to storage
		SubmitOrder(orderNum, orderDate, email, "Central Perk", main, side1, side2, drink, street1, street2, city, state, zipcode)

		//Display Operation Status Page to User
		orderStatus(w, r, statusData)
	}
}

func SubmitOrder(orderNum int, orderDate string, email string, restaurant string, main string, side1 string, side2 string, drink string, street1 string, street2 string, city string, state string, zipcode string) {

	//FOR TROUBLESHOOTING INPUT VALIDATION
	//fmt.Println("I'm begging the Submit Order Function Now - Trying to Submit to Kafka")
	//fmt.Println("#########")
	//fmt.Println("email is : " + email)
	//fmt.Println("restaurant is : " + restaurant)
	//fmt.Println("main is : " + main)
	//fmt.Println("side1 is : " + side1)
	//fmt.Println("side2 is : " + side2)
	//fmt.Println("drink is : " + drink)
	//fmt.Println("#########")

	//create a kafka producer - connection variables come from environment variables KAFKA_HOST and KAFKA_PORT
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaHost + ":" + kafkaPort})
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}

	//configure a kafka delivery channel
	deliveryChan := make(chan kafka.Event)

	// Produce messages to topic (asynchronously)
	topic := "order"
	msg := PxOrder{
		Email:       email,
		OrderId:     orderNum,
		Restaurant:  restaurant,
		Main:        main,
		Side1:       side1,
		Side2:       side2,
		Drink:       drink,
		Date:        orderDate,
		Street1:     street1,
		Street2:     street2,
		City:        city,
		State:       state,
		Zip:         zipcode,
		OrderStatus: "Pending",
	}

	//convert msg to a json object and store it in payload
	payload, err := json.Marshal(msg)

	//push msg to the kafka queue
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "PX-Delivery", Value: []byte("header values are binary")}},
	}, deliveryChan)
	if err != nil {
		fmt.Printf("Produce failed: %v\n", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)
}
