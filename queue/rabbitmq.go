
package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ() (*RabbitMQ, error) {

	
	conn, err := amqp.Dial(
		"amqp://guest:guest@localhost:5672/",
	)

	if err != nil {
		return nil, fmt.Errorf(
		"failed to connect to RabbitMQ: %w",
		err,
	)
	}

	
	channel, err := conn.Channel()

	if err != nil {
		conn.Close()

		return nil, fmt.Errorf(
			"failed to create RabbitMQ channel: %w",
			err,
		)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

func (r *RabbitMQ) Setup() error {

	
	err := r.channel.ExchangeDeclare(
		"task_exchange", // exchange name
		"direct",        // exchange type
		true,            // durable
		false,           // auto delete
		false,           // internal
		false,           // no wait
		nil,             // arguments
	)

	if err != nil {
		return fmt.Errorf(
			"failed to declare exchange: %w",
			err,
		)
	}

	// Create queue
	_, err = r.channel.QueueDeclare(
		"task_queue", // queue name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf(
			"failed to declare queue: %w",
			err,
		)
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		"task_queue",    // queue name
		"task.created",  // routing key
		"task_exchange", // exchange
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to bind queue: %w",
			err,
		)
	}

	return nil
}

// Close closes channel and connection
func (r *RabbitMQ) Close() error {

	if r.channel != nil {
		r.channel.Close()
	}

	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}

