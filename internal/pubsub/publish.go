package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	//* Convert the value to JSON
	body, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Failed to marshal JSON:", err)
		return err
	}

	//* Publish the message to the exchange with the specified routing key
	//* context.Background means empty context, no timeout or cancellation
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	return nil
}

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	// Create a channel
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the queue with properties based on queueType
	queue, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,   // durable
		queueType == Transient, // autoDelete
		queueType == Transient, // exclusive
		false,                  // noWait
		nil,                    // args
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare queue %q: %w", queueName, err)
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue %q to exchange %q with key %q: %w", queue.Name, exchange, key, err)
	}

	return channel, queue, nil
}
