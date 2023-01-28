package lib

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func KafkaCheck(KafkaHost string, KafkaPort string) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaHost + ":" + kafkaPort})
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connected to Kafka!")
		fmt.Println(p)
	}
}