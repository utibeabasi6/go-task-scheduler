package main

type RedisConfig struct {
	Url string `json:"url"`
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
