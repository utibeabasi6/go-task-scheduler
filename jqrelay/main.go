package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

var (
	VERSION = ""
	TOPIC   = ""
	PORT    = ""
	BROKERS = ""
)

func init() {
	flag.StringVar(&VERSION, "version", "0.11.0.0", "apache kafka version")
	flag.StringVar(&TOPIC, "topic", "task", "the topic to send messages to")
	flag.StringVar(&PORT, "port", "8000", "the port to start the server on")
	flag.StringVar(&BROKERS, "brokers", "localhost:9092", "comma seperated list of brokers to connect to")

	flag.Parse()
}

func main() {
	keepRunning := true
	var config Config = Config{Port: PORT, Kafka: KafkaConfig{Brokers: strings.Split(BROKERS, ","), Topic: TOPIC}}

	log.Println("Starting consumer for topic:", config.Kafka.Topic)

	// init consumer
	consumerConfig := sarama.NewConfig()
	consumer := Consumer{ready: make(chan bool, 1)}
	cg, err := sarama.NewConsumerGroup(config.Kafka.Brokers, "taskProcessors", consumerConfig)
	if err != nil {
		log.Fatalln("Error creating consumer group", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := cg.Consume(ctx, strings.Split(config.Kafka.Topic, ","), &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Println("Consumer running...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for keepRunning {
		select {
		case <-sigterm:
			log.Println("Sigterm recieved: exiting...")
			keepRunning = false
		case <-ctx.Done():
			log.Println("Context cancelled: exiting...")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()
	if err := cg.Close(); err != nil {
		log.Fatalln("Error closing consumer")
	}
}
