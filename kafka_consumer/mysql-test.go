package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Data struct {
	OrderId     int
	Email       string
	Main        string
	Side1       string
	Side2       string
	drink       string
	Restaurant  string
	Date        string
	Street1     string
	Street2     string
	City        string
	State       string
	Zipcode     string
	OrderStatus string
}

func writeToDB() {
	//open connection to mysql
	db, err := sql.Open("mysql", "root:porxie@tcp(127.0.0.1:3306)/")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Create database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE delivery")
	if err != nil {
		panic(err)
	}

	//change datbase
	_, err = db.Exec("USE delivery")
	if err != nil {
		panic(err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS orders (id INT AUTO_INCREMENT PRIMARY KEY, orderid int, email VARCHAR(255), main VARCHAR(255), side1 VARCHAR(255), side2 VARCHAR(255), drink VARCHAR(255), restaurant VARCHAR(255), date VARCHAR(255), street1 VARCHAR(255), street2 VARCHAR(255), city VARCHAR(255), state VARCHAR(255), zipcode VARCHAR(255), orderstatus VARCHAR(255))")
	if err != nil {
		panic(err.Error())
	}

}

func main() {

	//Push data to postgresql
	//print struct values

	fmt.Println("### Calling Write to DB Function ###")
	writeToDB()

}
