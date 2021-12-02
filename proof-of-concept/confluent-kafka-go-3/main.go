package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"fmt"
)

type ProducerImpl struct {
	producer *kafka.Producer
	topic    string
}

type ConsumerImpl struct {
	consumer *kafka.Consumer
}

type Order struct {
	Date    time.Time `json:"date"`
	Items   []string  `json:"items"`
	Payment string    `json:"payment"`
	Price   float64   `json:"price"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cancelChan := make(chan os.Signal)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	p, err := NewProducer("localhost:9092", "ORDERS")
	if err != nil {
		log.Fatal("cannot create producer: ", err)
	}

	c, err := NewConsumer("localhost:9092", "ORDER-PROCESSING-SERVICE")
	if err != nil {
		log.Fatal("cannot create order processing consumer: ", err)
	}

	// Produce
	go func(ctx context.Context) {
		for {
			time.Sleep(time.Second)
			order := Order{
				Date:    time.Now(),
				Items:   []string{"macbook", "iPhone", "iMac"},
				Payment: "CREDIT_CARD",
				Price:   rand.Float64() * 25000,
			}

			valBytes, err := json.Marshal(&order)
			if err != nil {
				log.Fatal("cannot json marshal value: ", order)
				cancel()
			}
			_, err = p.Send("user.1", valBytes)
			if err != nil {
				log.Fatal("cannot send message to broker: ", err)
				cancel()
			}
		}
	}(ctx)

	// Receive
	go func(ctx context.Context) {
		err := c.Listen([]string{"ORDERS"})
		if err != nil {
			log.Fatal(err)
			cancel()
		}
	}(ctx)

	for {
		select {
		case <-cancelChan:
			cancel()
		case <-ctx.Done():
			os.Exit(1)
		}
	}

}


// PRODUCER
func NewProducer(broker, topic string) (*ProducerImpl, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %s", err)
	}

	if topic == "" {
		return nil, fmt.Errorf("you need to specify an valid topic")
	}

	return &ProducerImpl{producer: p, topic: topic}, nil
}

func (p *ProducerImpl) Send(key string, val []byte, headers ...kafka.Header) (chan kafka.Event, error) {
	deliveryChan := make(chan kafka.Event)

	if err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          []byte(val),
		Headers:        headers,
	}, deliveryChan); err != nil {
		return nil, err
	}

	return deliveryChan, nil
}

// CONSUMER
func NewConsumer(broker, group string) (*ConsumerImpl, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               broker,
		"group.id":                        group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &ConsumerImpl{
		consumer: c,
	}, nil
}

func (c *ConsumerImpl) Listen(topics []string) error {
	defer c.consumer.Close()

	if err := c.consumer.SubscribeTopics(topics, nil); err != nil {
		return err
	}

	run := true
	for run {
		select {
		case ev := <-c.consumer.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				err := c.consumer.Assign(e.Partitions)
				if err != nil {
					log.Fatal(err)
				}
			case kafka.RevokedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				err := c.consumer.Unassign()
				if err != nil {
					log.Fatal(err)
				}
			case *kafka.Message:
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
		}
	}

	return nil
}
