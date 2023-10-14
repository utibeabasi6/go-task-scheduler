package main

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
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

type HttpResponse struct {
	Message string `json:"message"`
}

type CustomError struct {
	Error string `json:"error"`
}
