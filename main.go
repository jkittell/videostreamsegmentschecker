package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jkittell/data/database"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	jobDone := make(chan bool)
	jobs := make(chan Job)
	toCheck := make(chan Job)

	resultsDB, err := database.NewPostgresDB[*Result]("segment_checks", func() *Result {
		return &Result{}
	})
	failOnError(err, "unable to connect to postgres")

	infoDB, err := database.NewMongoDB[SegmentCheckInfo]()
	failOnError(err, "unable to connect to mongo")

	// receive url to check on q.segments.check.in
	go receiveStreamToCheck(jobs, jobDone)

	// send request for segments on q.segments.in
	go requestSegments(jobs)

	// get response for segments on q.segments.out
	go receiveSegments(toCheck)

	// check segments and store results in postgres
	go checkSegments(resultsDB, infoDB, toCheck, jobDone)

	router := gin.Default()
	router.GET("/api/:id", HandleSegmentCheckInfo(infoDB))
	err = router.Run(":8081")
	if err != nil {
		log.Fatal(err)
	}
}
