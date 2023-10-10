package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if _, err := os.Stat("config.json"); err != nil {
		log.Fatalln("Unable to locate config file", err)
	}

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Error while opening file", err)
	}

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln("Error while opening file", err)
	}

	http.HandleFunc("/job", func(http.ResponseWriter, *http.Request) {
		log.Println("Adding job...")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%v", config.Port)
	}

	log.Println("Starting listener on port", port)

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
