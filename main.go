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
	jobDone := make(chan bool)

	jobs := make(chan Job)
	toCheck := make(chan Job)

	// receive url to check on q.segments.check.in
	go receiveStreamToCheck(jobs, jobDone)

	// send request for segments on q.segments.in
	go requestSegments(jobs)

	// get response for segments on q.segments.out
	go receiveSegments(toCheck)

	// check segments and store results in postgres
	go checkSegments(toCheck, jobDone)
	<-done
}
