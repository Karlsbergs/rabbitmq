package server

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

var server *Server

type Server struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   map[string]amqp.Queue
}

func New() (*Server, error) {
	if server == nil {
		config, err := ReadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %s", err)
		}

		server = &Server{}

		url := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.UserName, config.Password, config.Host, config.Port)

		server.conn, err = amqp.Dial(url)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err)
		}

		err = server.CheckChannel()
		if err != nil {
			return nil, fmt.Errorf("failed to init channel: %s", err)
		}

		server.queue = make(map[string]amqp.Queue)

		log.Printf("Connected to %s", url)
	}

	return server, nil
}

func (s *Server) CheckChannel() error {
	if s.channel == nil {
		var err error
		s.channel, err = s.conn.Channel()
		if err != nil {
			return fmt.Errorf("failed to open channel: %s", err)
		}
	}

	return nil
}

func (s *Server) CheckQueue(name string) error {
	if _, ok := s.queue[name]; !ok {
		var err error
		s.queue[name], err = s.channel.QueueDeclare(
			name,  // name
			false, // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare a queue: %s", err)
		}
	}

	return nil
}

func (s *Server) Send(message SendMessage) error {
	err := message.Validate()
	if err != nil {
		return fmt.Errorf("message validation error: %s", err)
	}

	err = s.CheckQueue(message.Queue)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %s", err)
	}

	err = s.channel.Publish(
		"",            // exchange
		message.Queue, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message.Body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %s", err)
	}

	log.Printf("Sent: %s", message.Body)

	return nil
}

func (s *Server) Listen(queue string, listening chan bool, received chan amqp.Delivery) error {
	var err error

	err = s.CheckQueue(queue)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %s", err)
	}

	messages, err := s.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare a consumer: %s", err)
	}

	infinite := make(chan bool)

	go func() {
		for d := range messages {
			received <- d
		}
	}()

	log.Println("Listening for messages")

	// send `server listening` notification
	listening <- true

	<-infinite

	return nil
}

func (s *Server) Close() error {
	err := s.channel.Close()
	if err != nil {
		return fmt.Errorf("error closing channel: %s", err)
	}

	err = s.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %s", err)
	}

	log.Println("Server closed")

	return nil
}
