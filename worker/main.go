package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func processTask(val string, rdb *redis.Client) {
	var jobObj Job
	err := json.Unmarshal([]byte(val), &jobObj)
	if err != nil {
		log.Fatalln("Unable to decode job", err)
	}
	log.Println("processing job", jobObj.Id, "with payload", jobObj.Payload)
	log.Println("removing job from processing queue")
	rdb.LRem(context.TODO(), "processing", 0, val)
}

func main() {
	if _, err := os.Stat("config.json"); err != nil {
		log.Fatalln("Unable to locate config file", err)
	}

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Error while opening config file.", err)
	}

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while decoding config file", err)
	}

	workerType := os.Getenv("WORKER_TYPE")
	if workerType == "" {
		log.Fatalln("WORKER_TYPE env not set")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Url,
		Password: "", // no password set
		DB:       0,  // use default DB

	})

	for {
		data := rdb.LMove(context.TODO(), workerType, "processing", "LEFT", "RIGHT")
		val, err := data.Result()
		if err != nil {
			continue
		}
		go processTask(val, rdb)
	}

}