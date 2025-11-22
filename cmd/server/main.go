package main

import (
	"fmt"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn := "amqp://guest:guest@localhost:5672/"
	rabbitmqConn, err := amqp.Dial(conn)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer rabbitmqConn.Close()
	fmt.Println("Connection successful")
	fmt.Println("Starting Peril server...")

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Shutting down gracefully...")

}
