package main

import (
	"context"
	"flag"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	REDIS_URL        string
	TASK_QUEUE       string
	PROCESSING_QUEUE string
)

func init() {
	flag.StringVar(&REDIS_URL, "version", "localhost:6379", "redis url")
	flag.StringVar(&TASK_QUEUE, "task-queue", "task_queue", "queue where tasks are sent")
	flag.StringVar(&PROCESSING_QUEUE, "processing-queue", "processing_queue", "processing queue")
	flag.Parse()
}

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     REDIS_URL,
		Password: "", // no password set
		DB:       0,  // use default DB

	})

	log.Println("Starting worker...")
	for {
		data := rdb.LMove(context.TODO(), TASK_QUEUE, PROCESSING_QUEUE, "LEFT", "RIGHT")
		val, err := data.Result()
		if err != nil {
			continue
		}
		go processTask(val, rdb, PROCESSING_QUEUE)
	}

}
