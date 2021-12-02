package main

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ProducerImpl struct {
	producer *kafka.Producer
	topic    string
}

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
