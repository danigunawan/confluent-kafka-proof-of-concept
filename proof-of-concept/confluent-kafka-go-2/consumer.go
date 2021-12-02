package main

import (
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ConsumerImpl struct {
	consumer *kafka.Consumer
}

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
