package main

import (
	"bytes"
	"context"
	"encoding/gob"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"time"
)

func requestSegments(payloads chan Payload) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"segments.request", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for payload := range payloads {
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		if err := encoder.Encode(payload); err != nil {
			log.Println(err)
			continue
		}
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/x-gob",
				Body:        buffer.Bytes(),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf("[<<] Sent %s\n", payload.Id.String())
	}
}
