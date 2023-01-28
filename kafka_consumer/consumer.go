package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
)

type Data struct {
	OrderId     int
	Email       string
	Main        string
	Side1       string
	Side2       string
	Drink       string
	Restaurant  string
	Date        string
	Street1     string
	Street2     string
	City        string
	State       string
	Zipcode     string
	OrderStatus string
}

func ErrorCheck(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func writeToDB(payload Data) {

	//open connection to mysql
	db, err := sql.Open("mysql", "root:porxie@tcp(127.0.0.1:3306)/") //TODO CHANGE HARD CODING
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	//change datbase
	_, err = db.Exec("USE delivery")
	if err != nil {
		// Create database if it doesn't exist
		_, err = db.Exec("CREATE DATABASE delivery")
		if err != nil {
			println("Delivery database exists")
		}
		_, err = db.Exec("USE delivery")
	}

	// Create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS orders (id INT AUTO_INCREMENT PRIMARY KEY, orderid BIGINT, email VARCHAR(255), main VARCHAR(255), side1 VARCHAR(255), side2 VARCHAR(255), drink VARCHAR(255), restaurant VARCHAR(255), date VARCHAR(255), street1 TEXT, street2 TEXT, city VARCHAR(255), state VARCHAR(255), zipcode VARCHAR(255), orderstatus VARCHAR(255))")
	if err != nil {
		println("Order Table Exists")
	}

	//stmt := "INSERT INTO orders(orderid, email, main, side1, side2, drink, restaurant, date, street1, street2, city, state, zipcode, orderstatus) VALUES (" + "12345" + "," + "'bart@test.com'" + "," + payload.Main + "," + payload.Side1 + "," + payload.Side2 + "," + payload.Drink + "," + payload.Restaurant + "," + payload.Date + "," + "STREET1" + "," + "STREET2" + "," + payload.City + "," + payload.State + "," + payload.Zipcode + "," + payload.OrderStatus + ")"
	res, err := db.Exec("INSERT INTO orders(orderid, email, main, side1, side2, drink, restaurant, date, street1, street2, city, state, zipcode, orderstatus) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", payload.OrderId, payload.Email, payload.Main, payload.Side1, payload.Side2, payload.Drink, payload.Restaurant, payload.Date, payload.Street1, payload.Street2, payload.City, payload.State, payload.Zipcode, payload.OrderStatus)
	ErrorCheck(err)

	if err != nil {
		panic(err.Error())
	}

	lastId, err := res.LastInsertId()
	fmt.Printf("The last inserted row id: %d\n", lastId)
}

func main() {

	var (
		kafkaHost string = os.Getenv("KAFKA_HOST")
		kafkaPort string = os.Getenv("KAFKA_PORT")
	)

	conf := kafka.ReaderConfig{
		Brokers:  []string{kafkaHost + ":" + kafkaPort},
		Topic:    "order",
		GroupID:  "g1",
		MaxBytes: 10,
	}

	reader := kafka.NewReader(conf)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("Some error occured", err)
			continue
		}
		//Print all Messages retrieved - for testing
		//fmt.Println("message is : ", string(msg.Value))
		order := msg.Value

		//convert kafka messages to Struct
		var payload Data
		err = json.Unmarshal(order, &payload)
		if err != nil {
			log.Fatal("Error during Unmarshal(): ", err)
		}
		//Push data to postgresql
		//print struct values
		println("email is : " + payload.Email)

		fmt.Println("### Calling Write to DB Function ###")
		writeToDB(payload)

	}

}
