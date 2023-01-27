package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	//"github.com/confluentinc/confluent-kafka-go/schemaregistry"
	//"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde"
	//"github.com/confluentinc/confluent-kafka-go/schemaregistry/serde/jsonschema"
	"go.mongodb.org/mongo-driver/bson"
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
	OrderId    int    `bson:"orderid,omitempty"`
	Email      string `bson:"email,omitempty"`
	Restaurant string `bson:"restaurant,omitempty"`
	Date       string `bson:"date,omitempty"`
	Street1    string `bson:"street1,omitempty"`
	Street2    string `bson:"street2,omitempty"`
	City       string `bson:"city,omitempty"`
	State      string `bson:"state,omitempty"`
	Zip        string `bson:"zip,omitempty"`
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
			t, _ := template.ParseFiles("./static/pxbbq_order.html")
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

func McdOrderHandler(w http.ResponseWriter, r *http.Request) {
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
			t, _ := template.ParseFiles("./static/mcd_order.html")
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

func CentralperkOrderHandler(w http.ResponseWriter, r *http.Request) {
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
			t, _ := template.ParseFiles("./static/centralperk_order.html")
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

func SubmitOrder() {
	//url := "http://localhost:8081"
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:29092"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}

	fmt.Printf("Created Producer %v\n", p)

	//client, err := schemaregistry.NewClient(schemaregistry.NewConfig(url))

	//if err != nil {
	//	fmt.Printf("Failed to create schema registry client: %s\n", err)
	//}

	//ser, err := jsonschema.NewSerializer(client, serde.ValueSerde, jsonschema.NewSerializerConfig())

	//if err != nil {
	//	fmt.Printf("Failed to create serializer: %s\n", err)
	//}

	// Optional delivery channel, if not specified the Producer object's
	// .Events channel is used.
	deliveryChan := make(chan kafka.Event)

	// Produce messages to topic (asynchronously)
	topic := "order"
	msg := PxOrder{
		Email:      "eshanks@purestorage.com",
		OrderId:    12345,
		Restaurant: "pxbbq",
		Date:       "01-01-2023",
		Street1:    "123 main street",
		Street2:    "",
		City:       "springfield",
		State:      "IL",
		Zip:        "60606",
	}

	//testing
	payload, err := json.Marshal(msg)

	//end testing

	//payload, err := ser.Serialize(topic, &msg)
	//if err != nil {
	//	fmt.Printf("Failed to serialize payload: %s\n", err)
	//}

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
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
