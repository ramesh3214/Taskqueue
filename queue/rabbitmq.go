package queue

import (
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ() (*RabbitMQ, error) {

	rabbitURL := os.Getenv("RABBITMQ_URL")

	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
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

	// --------------------------------
	// Main Exchange
	// --------------------------------
	err := r.channel.ExchangeDeclare(
		"task_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare task exchange: %w",
			err,
		)
	}

	// --------------------------------
	// Retry Exchange
	// --------------------------------
	err = r.channel.ExchangeDeclare(
		"retry_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare retry exchange: %w",
			err,
		)
	}

	// --------------------------------
	// DLQ Exchange
	// --------------------------------
	err = r.channel.ExchangeDeclare(
		"dlq_exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare DLQ exchange: %w",
			err,
		)
	}

	// --------------------------------
	// Main Queue
	// --------------------------------
	_, err = r.channel.QueueDeclare(
		"task_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare task queue: %w",
			err,
		)
	}

	// --------------------------------
	// Retry Queue
	// --------------------------------
	_, err = r.channel.QueueDeclare(
		"task_retry_queue",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":          int32(5000),
			"x-dead-letter-exchange": "task_exchange",

			"x-dead-letter-routing-key": "task.created",
		},
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare retry queue: %w",
			err,
		)
	}

	// --------------------------------
	// DLQ
	// --------------------------------
	_, err = r.channel.QueueDeclare(
		"task_dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to declare DLQ: %w",
			err,
		)
	}

	// --------------------------------
	// Main Queue Binding
	// --------------------------------
	err = r.channel.QueueBind(
		"task_queue",
		"task.created",
		"task_exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to bind task queue: %w",
			err,
		)
	}

	// --------------------------------
	// Retry Queue Binding
	// --------------------------------
	err = r.channel.QueueBind(
		"task_retry_queue",
		"task.retry",
		"retry_exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to bind retry queue: %w",
			err,
		)
	}

	// --------------------------------
	// DLQ Binding
	// --------------------------------
	err = r.channel.QueueBind(
		"task_dlq",
		"task.failed",
		"dlq_exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to bind DLQ: %w",
			err,
		)
	}

	return nil
}

func (r *RabbitMQ) Publish(
	exchange string,
	routingKey string,
	message string,
) error {

	err := r.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         []byte(message),
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to publish message: %w",
			err,
		)
	}

	return nil
}

func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {

	messages, err := r.channel.Consume(
		"task_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to register consumer: %w",
			err,
		)
	}

	return messages, nil
}

func (r *RabbitMQ) Close() error {

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}

	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}
