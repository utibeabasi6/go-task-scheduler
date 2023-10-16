package main

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}
type Config struct {
	Port  string      `json:"port"`
	Kafka KafkaConfig `json:"kafkaConfig"`
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
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Println("Message channel was closed")
				return nil
			}
			log.Println("consumed message", string(msg.Value))
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
