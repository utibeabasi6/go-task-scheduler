package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	if _, err := os.Stat("config.json"); err != nil {
		log.Fatalln("Unable to locate config file", err)
	}

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Error while opening file.", err)
	}

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while opening file", err)
	}

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		var message HttpResponse
		var jobObj Job

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&jobObj)
		if err != nil {
			message.Message = "Unable to deserialize json"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(message)
			return
		}
		if jobObj.Payload == "" || jobObj.Type == "" {
			message.Message = "Unable to deserialize data, one of jobId, jobType or payload not set"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(message)
			return
		}

		jobObj.Id = hash(jobObj.Type, jobObj.Payload)

		rdb := redis.NewClient(&redis.Options{
			Addr:     config.Redis.hosts[jobObj.Type],
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		// Deduplication logic
		elementRequest := rdb.LRange(context.TODO(), jobObj.Type, 0, -1)
		elements, err := elementRequest.Result()
		if err != nil {
			message.Message = "Unable to fetch all objects from redis"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(message)
			return
		}
		var deduplicatedObj Job
		for _, v := range elements {
			err := json.Unmarshal([]byte(v), &deduplicatedObj)
			if err != nil {
				message.Message = "Unable to deserialize redis obj"
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(message)
				return
			}
			if deduplicatedObj.Id == jobObj.Id {
				message.Message = "Unable to add to queue. Object already exists"
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(message)
				return
			}
		}

		encoded, err := json.Marshal(jobObj)
		if err != nil {
			message.Message = "Unable to encode json"
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(message)
			return
		}

		cmd := rdb.LPush(context.TODO(), jobObj.Type, encoded)
		if err := cmd.Err(); err != nil {
			message.Message = fmt.Sprintf("An error occured while pushing job %s to queue: %v", jobObj.Id, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(message)
			return
		}

		message.Message = fmt.Sprintf("Job %s added to queue", jobObj.Id)
		log.Println(message.Message)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(message)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%v", config.Port)
	}

	log.Println("Starting listener on port", port)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
