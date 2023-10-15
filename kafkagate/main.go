package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/IBM/sarama"
)

var (
	VERSION = ""
	TOPIC   = ""
	PORT    = ""
	BROKERS = ""
)

func init() {
	flag.StringVar(&VERSION, "version", "0.11.0.0", "apache kafka version")
	flag.StringVar(&TOPIC, "topic", "task", "the topic to send messages to")
	flag.StringVar(&PORT, "port", "8000", "the port to start the server on")
	flag.StringVar(&BROKERS, "brokers", "localhost:9092", "comma seperated list of brokers to connect to")

	flag.Parse()
}

func main() {
	var config Config = Config{Port: PORT, Kafka: KafkaConfig{Brokers: strings.Split(BROKERS, ","), Topic: TOPIC}}

	// Create kafka producer
	producerConfg := sarama.NewConfig()
	producerConfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(config.Kafka.Brokers, producerConfg)
	if err != nil {
		log.Fatalln("Error while creating kafka producer.", err)
	}
	clusterAdmin, err := sarama.NewClusterAdmin(config.Kafka.Brokers, producerConfg)
	if err != nil {
		log.Fatalln("Error while creating kafka cluster admin.", err)
	}
	numBrokers := len(config.Kafka.Brokers)
	err = clusterAdmin.CreateTopic(config.Kafka.Topic, &sarama.TopicDetail{
		ReplicationFactor: int16(numBrokers),
		NumPartitions:     1,
	}, false)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln("Unable to close producer")
		}
		if err := clusterAdmin.Close(); err != nil {
			log.Fatalln("Unable to close cluster admin")
		}
	}()

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		var message HttpResponse
		var jobObj Job

		w.Header().Set("Content-Type", "application/json")

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&jobObj)
		if err != nil {
			errMessage := fmt.Sprintf("Unable to encode json. %v", err)
			handleErrors(w, errMessage, http.StatusInternalServerError)
			return
		}
		if jobObj.Payload == "" || jobObj.Type == "" {
			errMessage := "Unable to deserialize data, one of jobId, jobType or payload not set"
			handleErrors(w, errMessage, http.StatusBadRequest)
			return
		}

		// Send job to kafka topic
		jobObj.Id = hash(jobObj.Type, jobObj.Payload)
		bytes, err := json.Marshal(jobObj)
		if err != nil {
			errMessage := fmt.Sprintf("Unable to encode json. %v", err)
			handleErrors(w, errMessage, http.StatusInternalServerError)
			return
		}
		producerMessage := sarama.ProducerMessage{Topic: TOPIC, Value: sarama.StringEncoder(string(bytes))}
		partition, offset, err := producer.SendMessage(&producerMessage)
		if err != nil {
			errMessage := fmt.Sprintf("Error sending job to kafka. %v", err)
			handleErrors(w, errMessage, http.StatusInternalServerError)
			return
		}

		message.Message = fmt.Sprintf("Job %s added to kafka, partition: %d, offset: %d", jobObj.Id, partition, offset)
		log.Println(message.Message)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(message)

		// rdb := redis.NewClient(&redis.Options{
		// 	Addr:     config.Redis.Url,
		// 	Password: "", // no password set
		// 	DB:       0,  // use default DB
		// })

		// Deduplication logic
		// encoded, err := json.Marshal(jobObj)
		// if err != nil {
		// 	errMessage := fmt.Sprintf("Unable to encode json. %v", err)
		// 	handleErrors(w, errMessage, http.StatusInternalServerError)
		// 	return
		// }

		// check if object is in "queue" queue
		// elementPos := rdb.LPos(context.TODO(), "queue", string(encoded), redis.LPosArgs{})
		// if err := elementPos.Err(); err == nil {
		// 	errMessage := fmt.Sprintf("Unable to add to queue. Object already exists. %v", err)
		// 	handleErrors(w, errMessage, http.StatusBadRequest)
		// 	return
		// }

		// check if object is in "processing" queue
		// elementPos = rdb.LPos(context.TODO(), "processing", string(encoded), redis.LPosArgs{})
		// if err := elementPos.Err(); err == nil {
		// 	errMessage := fmt.Sprintf("Unable to add to queue. Object being processed. %v", err)
		// 	handleErrors(w, errMessage, http.StatusBadRequest)
		// 	return
		// }

		// Push job to redis
		// cmd := rdb.LPush(context.TODO(), "queue", encoded)
		// if err := cmd.Err(); err != nil {
		// 	errMessage := fmt.Sprintf("An error occured while pushing job %s to queue: %v", jobObj.Id, err)
		// 	handleErrors(w, errMessage, http.StatusInternalServerError)
		// 	return
		// }
	})

	port := config.Port
	if port == "" {
		port = fmt.Sprintf("%v", config.Port)
	}

	log.Println("Starting listener on port", port)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
