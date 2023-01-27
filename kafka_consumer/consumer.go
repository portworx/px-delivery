package main

import (
	"context"
	"fmt"
	"os"

	"github.com/segmentio/kafka-go"
)

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
		fmt.Println("message is : ", string(msg.Value))
	}

}
