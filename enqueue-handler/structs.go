package main

type RedisConfig struct {
	hosts map[string]string
}
type Config struct {
	Port  int         `json:"port"`
	Redis RedisConfig `json:"redisConfig"`
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
