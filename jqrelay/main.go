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
	consulapi "github.com/hashicorp/consul/api"
)

var (
	TOPIC      = ""
	PORT       = ""
	BROKERS    = ""
	REDIS_URL  = ""
	CONSUL_URL = ""
)

func init() {
	flag.StringVar(&TOPIC, "topic", "task", "the topic to send messages to")
	flag.StringVar(&PORT, "port", "8000", "the port to start the server on")
	flag.StringVar(&BROKERS, "brokers", "localhost:9092", "comma seperated list of brokers to connect to")
	flag.StringVar(&TASK_QUEUE, "task-queue", "task_queue", "the task queue")
	flag.StringVar(&PROCESSING_QUEUE, "processing-queue", "processing_queue", "the processing queue")
	flag.StringVar(&REDIS_URL, "redis-url", "localhost:6379", "the redis url")
	flag.StringVar(&CONSUL_URL, "consul-url", "localhost:8500", "the consul url")

	flag.Parse()
}

func main() {
	keepRunning := true
	JqRelayConfig = Config{Kafka: KafkaConfig{Brokers: strings.Split(BROKERS, ","), Topic: TOPIC}, Redis: RedisConfig{Url: REDIS_URL}, Consul: ConsulConfig{Url: CONSUL_URL}}
	log.Println("Starting consumer for topic:", JqRelayConfig.Kafka.Topic)

	// init consumer
	consumerConfig := sarama.NewConfig()
	consumer := Consumer{ready: make(chan bool, 1)}
	cg, err := sarama.NewConsumerGroup(JqRelayConfig.Kafka.Brokers, "taskProcessors", consumerConfig)
	if err != nil {
		log.Fatalln("Error creating consumer group", err)
	}

	// init consul client
	consulConfig := consulapi.Config{Address: JqRelayConfig.Consul.Url}
	consulClient, err := consulapi.NewClient(&consulConfig)

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Attempt to acquire consul lock
		log.Println("Acquiring consul lock...")
		lock, err := consulClient.LockKey(JqRelayConfig.Kafka.Topic)
		if err != nil {
			log.Fatalln("Unable to creating lock key", err)
		}
		lockChan, err := lock.Lock(nil)
		if err != nil {
			log.Fatalln("Unable to creating acquiring lock", err)
		}
		defer func() {
			log.Println("Releasing consul lock...")
			if err := lock.Unlock(); err != nil {
				log.Fatalln("Error occured while releasing lock", err)
			}
		}()
		// Process messages
		for {
			if err := cg.Consume(ctx, strings.Split(JqRelayConfig.Kafka.Topic, ","), &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
			select {
			case <-lockChan:
				log.Println("Consul lock lost prematurely")
				cancel()
			}
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
