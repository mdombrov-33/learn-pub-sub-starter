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

	// Connect to RabbitMQ
	rabbitmqConn, err := amqp.Dial(conn)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer rabbitmqConn.Close()

	// Ask for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Declare and bind a transient, exclusive queue for this client
	channel, _, err := pubsub.DeclareAndBind(
		rabbitmqConn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Println("Failed to declare and bind queue:", err)
		return
	}
	defer channel.Close()

	fmt.Println("Starting Peril client...")
	fmt.Println("Connection successful")
	fmt.Println("Specify your command")

	// Create a new game state for this client
	state := gamelogic.NewGameState(username)

	// Subscribe to pause messages
	err = pubsub.SubscribeJSON(rabbitmqConn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient, handlePause(state))
	if err != nil {
		fmt.Println("Failed to subscribe to pause messages:", err)
		return
	}

	/// Each game client should subscribe to moves from other players
	err = pubsub.SubscribeJSON(rabbitmqConn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.Transient, handleArmyMove(state))
	if err != nil {
		fmt.Println("Failed to subscribe to army moves:", err)
		return
	}

	// Setup graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// REPL loop
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "spawn":
			state.CommandSpawn(input)

		case "move":
			state.CommandMove(input)
			move, err := state.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Publish move
			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, string(routing.ArmyMovesPrefix)+"."+username, move)
			if err != nil {
				fmt.Println("Failed to publish move:", err)
				continue
			}
			fmt.Println("Move published successfully")

		case "status":
			state.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}

		// Non-blocking check for CTRL+C
		select {
		case <-signalChan:
			fmt.Println("Shutting down gracefully...")
			return
		default:
			// continue the loop
		}
	}
}

func handlePause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(state routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handleArmyMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutComeSafe, gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard

		default:
			return pubsub.NackDiscard
		}
	}
}
