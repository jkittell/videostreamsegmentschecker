package main

import (
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	done := make(chan bool)

	requests := make(chan Payload)
	streamsToCheck := make(chan Payload)

	// receive url to check on q.segments.checker.request
	go receiveStreamToCheck(requests)

	// send request for segments on segments.request
	go requestSegments(requests)

	// get response for segments on segments.response
	go receiveSegments(streamsToCheck)

	// check segments and store results in postgres
	go checkSegments(streamsToCheck)
	<-done
}
