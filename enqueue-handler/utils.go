package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
)

func hash(string1, string2 string) string {
	hasher := sha256.New()
	hasher.Write([]byte(string1 + string2))
	hashedString := hex.EncodeToString(hasher.Sum(nil))
	return hashedString
}

func handleErrors(w http.ResponseWriter, message string, status int) {
	errObj := CustomError{Error: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(errObj)
	if err != nil {
		log.Println(err)
	}
}
