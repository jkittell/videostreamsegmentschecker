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
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for range time.Tick(time.Second * 15) {
		payload := Job{
			Id:        uuid.New(),
			URL:       url,
			Segments:  nil,
			CreatedAt: time.Now(),
		}
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		if err := encoder.Encode(payload); err != nil {
			log.Println(err)
			continue
		}
		err = ch.PublishWithContext(ctx,
			"content", // exchange
			"",        // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/x-gob",
				Body:        buffer.Bytes(),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf("[<<] Sent segment check request: %s\n", payload.Id.String())
	}
}
