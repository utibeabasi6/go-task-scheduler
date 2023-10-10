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
		log.Fatalln("Error while opening config file.", err)
	}

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while decoding config file", err)
	}

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		var message HttpResponse
		var jobObj Job

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&jobObj)
		if err != nil {
			errMessage := "Unable to encode json"
			handleErrors(w, errMessage, http.StatusInternalServerError, err)
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
			Addr:     config.Redis.Url,
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		// Deduplication logic
		encoded, err := json.Marshal(jobObj)
		if err != nil {
			errMessage := "Unable to encode json"
			handleErrors(w, errMessage, http.StatusInternalServerError, err)
			return
		}
		// check if object is in "type" queue
		elementPos := rdb.LPos(context.TODO(), jobObj.Type, string(encoded), redis.LPosArgs{})
		if err := elementPos.Err(); err == nil {
			errMessage := fmt.Sprintf("Unable to add to queue. Object already exists. %v", err)
			handleErrors(w, errMessage, http.StatusBadRequest, err)
			return
		}

		// check if object is in "processing" queue
		elementPos = rdb.LPos(context.TODO(), "processing", string(encoded), redis.LPosArgs{})
		if err := elementPos.Err(); err == nil {
			errMessage := fmt.Sprintf("Unable to add to queue. Object being processed. %v", err)
			handleErrors(w, errMessage, http.StatusInternalServerError, err)
			return
		}

		// Push job to queue
		cmd := rdb.LPush(context.TODO(), "queue", encoded)
		if err := cmd.Err(); err != nil {
			errMessage := fmt.Sprintf("An error occured while pushing job %s to queue: %v", jobObj.Id, err)
			handleErrors(w, errMessage, http.StatusInternalServerError, err)
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
