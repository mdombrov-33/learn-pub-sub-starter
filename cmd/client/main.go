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

	rabbitmqConn, err := amqp.Dial(conn)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}

	channel, _, err := pubsub.DeclareAndBind(rabbitmqConn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
	if err != nil {
		fmt.Println("Failed to declare and bind queue:", err)
		return
	}

	defer channel.Close()
	defer rabbitmqConn.Close()

	fmt.Println("Starting Peril client...")
	fmt.Println("Connection successful")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	fmt.Println("Client running. Press CTRL+C to exit...")
	<-signalChan
	fmt.Println("Shutting down gracefully...")
}
