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

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {

	// 1. Ensure queue exists and binding is done
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("queue can't be bound to the exchange %q with key %q: %w", exchange, key, err)
	}

	// 2. Start consuming messages
	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to register a consumer for queue %q: %w", queue.Name, err)
	}

	// 3. Process messages in a goroutine
	go func() {
		for msg := range msgs {

			var payload T
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				fmt.Println("Failed to unmarshal JSON:", err)
				fmt.Println("[NACK:discard] due to unmarshal error")
				msg.Nack(false, false)
				continue
			}

			ack := handler(payload)

			switch ack {
			case Ack:
				fmt.Println("[ACK] message processed")
				msg.Ack(false)

			case NackRequeue:
				fmt.Println("[NACK:requeue] handler requested retry")
				msg.Nack(false, true)

			case NackDiscard:
				fmt.Println("[NACK:discard] handler requested drop")
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
