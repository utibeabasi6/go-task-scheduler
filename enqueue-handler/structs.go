package main

type RedisConfig struct {
	hosts map[string]string
}
type Config struct {
	Port  int         `json:"port"`
	Redis RedisConfig `json:"redisConfig"`
}
