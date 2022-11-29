package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"rabbitmq/internal/server"

	"github.com/streadway/amqp"
)

const queueName = "Test-Queue"
const sendMessageCount = 50

func main() {
	rServer, err := server.New()
	if err != nil {
		log.Fatalf("Connection Error: %s", err)
	}
	defer func() {
		err = rServer.Close()
		if err != nil {
			log.Fatalf("Connection Close Error: %s", err)
		}
	}()

	listening := make(chan bool)
	received := make(chan amqp.Delivery, 100)

	go rServer.Listen(queueName, listening, received)

	// wait for `server listening` notification before continuing
	<-listening

	go func() {
		for {
			message := <-received

			if ProcessMessage(message.Body) {
				// acknowlegde message processed
				message.Ack(false)
			}
		}
	}()

	// infinite loop sending messages to queue
	go func(max int) {
		for i := 0; i < max; i++ {
			err := rServer.Send(server.SendMessage{
				Queue: queueName, // + "1",
				Body:  fmt.Sprintf("Test-Message1-%d", i),
			})
			if err != nil {
				log.Printf("Send Error: %s", err)
				break
			}

			// slow sending to 1 message every 50 milliseconds
			time.Sleep(50 * time.Millisecond)
		}
	}(sendMessageCount)

	log.Println("CTRL+C to exit")

	// exit when we get a CTRL+C
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}

func ProcessMessage(body []byte) bool {
	log.Printf("Received: %s", string(body))

	// slow message processing to 1 message per 500 milliseconds
	time.Sleep(500 * time.Millisecond)

	return true
}
