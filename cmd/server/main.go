package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const conn = "amqp://guest:guest@localhost:5672/"

	//* TCP connection between the server and RabbitMQ instance
	rabbitmqConn, err := amqp.Dial(conn)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	//* Lightweight connection inside TCP connection for AMQP operations.
	//* RabbitMQ uses channels so it can do multiple things concurrently over a single TCP connection.
	channel, err := rabbitmqConn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}

	state := routing.PlayingState{
		IsPaused: true,
	}

	err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, state)
	if err != nil {
		fmt.Println("Failed to publish message:", err)
		return
	}

	defer channel.Close()
	defer rabbitmqConn.Close()

	fmt.Println("Starting Peril server...")
	fmt.Println("Connection successful")

	//* Handle graceful shutdown(CTRL+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Shutting down gracefully...")

}
