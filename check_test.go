package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"testing"
	"time"
)

var url = "http://devimages.apple.com/iphone/samples/bipbop/bipbopall.m3u8"

func TestCheck(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for range time.Tick(time.Second) {
		payload := Payload{
			Id:       uuid.New(),
			URL:      url,
			Segments: nil,
		}
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
