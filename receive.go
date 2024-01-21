package main

import (
	"bytes"
	"encoding/gob"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func receiveStreamToCheck(jobs chan Job, jobDone chan bool) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"content", // name
		"fanout",  // type
		false,     // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		"content", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	done := make(chan bool)

	go func() {
		for d := range messages {
			dec := gob.NewDecoder(bytes.NewReader(d.Body))
			var j Job
			err = dec.Decode(&j)
			if err != nil {
				log.Printf("decode message error: %s", err)
			}
			log.Printf("[ %s ] [>>] received segment check request", j.Id)
			jobs <- j
			log.Printf("[ %s ] [<-] waiting for segments", j.Id)
			<-jobDone
			log.Printf("[ %s ] [->] segment check completed", j.Id)
			d.Ack(false)
		}
	}()
	<-done
}

func receiveSegments(streamsToCheck chan Job) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"q.segments.response", // name
		false,                 // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	done := make(chan bool)

	go func() {
		for d := range messages {
			dec := gob.NewDecoder(bytes.NewReader(d.Body))
			var j Job
			err = dec.Decode(&j)
			if err != nil {
				log.Printf("[ %s ] decode message error: %s", j.Id, err)
			}
			log.Printf("[ %s ] [>>] received segments", j.Id)
			// check segments and write to database
			streamsToCheck <- j
			d.Ack(false)
		}
	}()
	<-done
}
