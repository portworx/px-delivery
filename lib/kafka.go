package lib

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	networkTCP = "tcp"
	partition  = 0
	topicName  = "order"
)

type config struct {
	User        string        `env:"KAFKA_USER" envDefault:"pds"`
	Password    string        `env:"KAFKA_PASS,required"`
	Host        string        `env:"KAKFA_HOST,required"`
	Port        string        `env:"PORT" envDefault:"9092"`
	Count       int           `env:"COUNT" envDefault:"100"`
	SleepTime   time.Duration `env:"SLEEP_TIME" envDefault:"5s"`
	Iterations  int           `env:"ITERATIONS" envDefault:"0"`
	FailOnError bool          `env:"FAIL_ON_ERROR" envDefault:"false"`
}

func KafkaInit() {
	// Read config.
	fmt.Println("Checking Kafka Broker")
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("ERROR: failed to parse config: %v\n", err)
	}

	brokerURL := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	dialer := &kafka.Dialer{
		SASLMechanism: plainMechanism(cfg.User, cfg.Password),
		Timeout:       10 * time.Second,
		DualStack:     true,
	}

	// Create topic.
	err := createTopic(dialer, brokerURL, topicName)
	if err != nil {
		return
	}

	// Wait few seconds to sync the topic and to avoid the "Unknown Topic Or Partition" error.
	time.Sleep(5 * time.Second)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

}

func connectToController(dialer *kafka.Dialer, url string) (*kafka.Conn, *kafka.Conn, error) {
	ctx := context.Background()

	// Connecting to broker url.
	//log.Printf("Connecting to %s\n", url)
	conn, err := dialer.DialContext(ctx, networkTCP, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open a connection: %s", err)
	}

	// Connecting to controller.
	controller, err := conn.Controller()
	if err != nil {
		return conn, nil, fmt.Errorf("failed to get the current controller: %s", err)
	}
	controllerUrl := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	//log.Printf("Connecting to controller %s\n", controllerUrl)
	controllerConn, err := dialer.DialContext(ctx, networkTCP, controllerUrl)
	if err != nil {
		return conn, nil, fmt.Errorf("failed to open a connection to the controller: %s", err)
	}

	return conn, controllerConn, err
}

func createTopic(dialer *kafka.Dialer, brokerURL, topicName string) error {
	// Connect to controller.
	conn, controllerConn, err := connectToController(dialer, brokerURL)
	if conn != nil {
		defer conn.Close()
	}
	if controllerConn != nil {
		defer controllerConn.Close()
	}
	if err != nil {
		return err
	}

	// Create topic.
	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topicName,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create the %s topic: %s", topicName, err)
	}
	fmt.Println("Kafka Topic " + topicName + " is ready!")

	return nil
}

func deleteTopic(dialer *kafka.Dialer, brokerURL, topicName string) error {
	// Connect to controller.
	conn, controllerConn, err := connectToController(dialer, brokerURL)
	if conn != nil {
		defer conn.Close()
	}
	if controllerConn != nil {
		defer controllerConn.Close()
	}
	if err != nil {
		return err
	}

	// Delete topic.
	err = controllerConn.DeleteTopics(topicName)
	if err != nil {
		return fmt.Errorf("failed to delete the %s topic: %s", topicName, err)
	}

	return nil
}

func writeMessages(dialer *kafka.Dialer, url string, topic string, count int) error {
	// Find leader node for topic.
	ctx := context.Background()
	leader, err := dialer.LookupLeader(ctx, networkTCP, url, topic, partition)
	if err != nil {
		return fmt.Errorf("failed to find leader for topic %s: %v", topic, err)
	}
	leaderURL := net.JoinHostPort(leader.Host, strconv.Itoa(leader.Port))

	log.Printf("write messages (%s)\n", leaderURL)
	w := newWriter(leaderURL, topic, dialer)
	defer w.Close()
	start := time.Now()
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key%d-%d", time.Now().Nanosecond(), i)
		err := w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(key),
			Value: []byte("this is message" + key),
		})
		if err != nil {
			if writeErr, ok := err.(kafka.WriteErrors); ok {
				for _, werr := range writeErr {
					log.Printf("ERROR: write #%d failed: %v\n", i, werr)
				}
			}
			return fmt.Errorf("write failed: %v", err)
		}
	}
	stop := time.Now()
	log.Printf("%d writes done in %v\n", count, stop.Sub(start))
	return nil
}

func readMessages(dialer *kafka.Dialer, url, topic string, count int) error {
	// Find leader node for topic.
	ctx := context.Background()
	leader, err := dialer.LookupLeader(ctx, networkTCP, url, topic, partition)
	if err != nil {
		return fmt.Errorf("failed to find leader for topic %s: %v", topic, err)
	}
	leaderURL := net.JoinHostPort(leader.Host, strconv.Itoa(leader.Port))

	log.Printf("read messages (%s)\n", leaderURL)
	r := newReader(leaderURL, topic, partition, dialer)
	defer r.Close()
	start := time.Now()
	for i := 0; i < count; i++ {
		_, err := r.ReadMessage(context.Background())
		if err != nil {
			return fmt.Errorf("read #%d failed: %v", i, err)
		}
	}
	stop := time.Now()
	log.Printf("%d reads done in %v\n", count, stop.Sub(start))
	return nil
}

func plainMechanism(user, password string) sasl.Mechanism {
	return plain.Mechanism{
		Username: user,
		Password: password,
	}
}

func newWriter(url string, topic string, dialer *kafka.Dialer) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(url),
		Topic:    topic,
		Balancer: &kafka.Hash{},
		Transport: &kafka.Transport{
			SASL: dialer.SASLMechanism,
		},
		BatchTimeout: 20 * time.Millisecond,
	}
}

func newReader(url string, topic string, partition int, dialer *kafka.Dialer) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{url},
		Topic:     topic,
		Dialer:    dialer,
		Partition: partition,
	})
}

func getTopicName() string {
	return fmt.Sprintf("loadtest%d", time.Now().Nanosecond())
}
