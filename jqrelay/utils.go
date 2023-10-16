package main

import (
	"context"
	"errors"
	"log"

	"github.com/redis/go-redis/v9"
)

func sendToRedis(rdb *redis.Client, msg, task_queue, processing_queue string) error {
	// Deduplication logic
	elementPos := rdb.LPos(context.TODO(), task_queue, msg, redis.LPosArgs{})
	if err := elementPos.Err(); err == nil {
		return errors.New(msg + " already exists in " + task_queue)
	}

	// check if object is in "processing" queue
	elementPos = rdb.LPos(context.TODO(), processing_queue, msg, redis.LPosArgs{})
	if err := elementPos.Err(); err == nil {
		return errors.New(msg + "already exists in " + processing_queue)
	}

	// Push job to redis
	log.Println("Pushing job", msg, "to redis")
	cmd := rdb.LPush(context.TODO(), task_queue, msg)
	if err := cmd.Err(); err != nil {
		return errors.New("Error occured while sending message to " + task_queue + ": " + err.Error())
	}
	return nil
}
