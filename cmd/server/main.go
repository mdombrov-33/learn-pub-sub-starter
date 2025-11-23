package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	channel, _, err := pubsub.DeclareAndBind(rabbitmqConn, routing.ExchangePerilTopic, "game_logs", "game_logs.*", pubsub.Durable)
	if err != nil {
		fmt.Println("Failed to declare and bind queue:", err)
		return
	}

	defer channel.Close()
	defer rabbitmqConn.Close()

	fmt.Println("Starting Peril server...")
	fmt.Println("Connection successful")
	gamelogic.PrintServerHelp()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		fmt.Println("\nShutting down gracefully...")
		channel.Close()
		rabbitmqConn.Close()
		os.Exit(0)
	}()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		command := input[0]
		switch command {
		case "pause":
			fmt.Println("Sending pause message")
			state := routing.PlayingState{IsPaused: true}
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, state)
			if err != nil {
				fmt.Println("Failed to publish message:", err)
				return
			}

		case "resume":
			fmt.Println("Sending resume message")
			state := routing.PlayingState{IsPaused: false}
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, state)
			if err != nil {
				fmt.Println("Failed to publish message:", err)
				return
			}

		case "quit":
			fmt.Println("Quitting server...")
			return

		default:
			fmt.Println("Unknown command")
		}
	}

}
