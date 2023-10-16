package main

import (
	"log"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Url string `json:"url"`
}

type ConsulConfig struct {
	Url string `json:"url"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}
type Config struct {
	Port   string       `json:"port"`
	Kafka  KafkaConfig  `json:"kafkaConfig"`
	Redis  RedisConfig  `json:"redisConfig"`
	Consul ConsulConfig `json:"consulConfig"`
}

type Job struct {
	Id      string `json:"jobId"`
	Type    string `json:"jobType"`
	Payload string `json:"payload"`
}

type Consumer struct {
	ready chan bool
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     JqRelayConfig.Redis.Url,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Println("Message channel was closed")
				return nil
			}
			err := sendToRedis(rdb, string(msg.Value), TASK_QUEUE, PROCESSING_QUEUE)
			if err != nil {
				log.Println("Error sending message to redis", err)
			} else {
				// Mark the message as processed
				session.MarkMessage(msg, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}
