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
)

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
