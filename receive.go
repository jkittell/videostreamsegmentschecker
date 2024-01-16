package main

import (
	"bytes"
	"encoding/gob"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
)

func receiveStreamToCheck(requests chan Payload) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"q.segments.checker.request", // name
		false,                        // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
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
			var p Payload
			err = dec.Decode(&p)
			if err != nil {
				log.Fatal("decode:", err)
			}
			log.Printf(" [>>] Received: %s\n", p.Id)
			requests <- p
			d.Ack(false)
		}
	}()

	<-done
}

func receiveSegments(streamsToCheck chan Payload) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"segments.response", // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
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
			var p Payload
			err = dec.Decode(&p)
			if err != nil {
				log.Fatal("decode:", err)
			}
			log.Printf("[>>] Received: %s\n", p.Id.String())
			// check segments and write to database
			streamsToCheck <- p
			d.Ack(false)
		}
	}()

	<-done
}
