package main

type Job struct {
	Id      string `json:"jobId"`
	Type    string `json:"jobType"`
	Payload string `json:"payload"`
}
