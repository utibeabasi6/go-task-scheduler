package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

func processTask(val string, rdb *redis.Client, processing_queue string) {
	var jobObj Job
	err := json.Unmarshal([]byte(val), &jobObj)
	if err != nil {
		log.Fatalln("Unable to decode job", err)
	}
	log.Println("processing job", jobObj.Id, "with payload", jobObj.Payload)
	log.Println("removing job from processing queue")
	rdb.LRem(context.TODO(), processing_queue, 0, val)
}
