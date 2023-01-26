package lib

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

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
	t, _ := template.ParseFiles("./static/internal_order_status.html")
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

	//if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
	//	fmt.Println("Not Authenticated")
	//} else {
	//	fmt.Println("Authenticated")
	//}

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
