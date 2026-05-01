package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageQueue provides a simple interface for message queue operations
type MessageQueue struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	url        string
}

// Message represents a message to be sent/received
type Message struct {
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

// NewMessageQueue creates a new message queue connection
func NewMessageQueue(url string) (*MessageQueue, error) {
	if url == "" {
		url = os.Getenv("RABBITMQ_URL")
		if url == "" {
			url = "amqp://guest:guest@localhost:5672/"
		}
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	mq := &MessageQueue{
		connection: conn,
		channel:    ch,
		url:        url,
	}

	return mq, nil
}

// Close closes the connection to the message queue
func (mq *MessageQueue) Close() error {
	if mq.channel != nil {
		mq.channel.Close()
	}
	if mq.connection != nil {
		return mq.connection.Close()
	}
	return nil
}

// DeclareQueue declares a queue
func (mq *MessageQueue) DeclareQueue(name string) error {
	_, err := mq.channel.QueueDeclare(
		name,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", name, err)
	}
	return nil
}

// Publish publishes a message to a queue
func (mq *MessageQueue) Publish(ctx context.Context, queue string, msg Message) error {
	msg.Timestamp = time.Now()

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use provided context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = mq.channel.PublishWithContext(
		timeoutCtx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (mq *MessageQueue) Consume(queue string, handler func(Message) error) error {
	msgs, err := mq.channel.Consume(
		queue,
		"",    // consumerTag
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for msg := range msgs {
		var mqMsg Message
		if err := json.Unmarshal(msg.Body, &mqMsg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			msg.Nack(false, false)
			continue
		}

		if err := handler(mqMsg); err != nil {
			log.Printf("Handler error: %v", err)
			msg.Nack(false, true) // requeue on error
			continue
		}

		msg.Ack(false)
	}

	return nil
}

// PublishToExchange publishes a message to an exchange
func (mq *MessageQueue) PublishToExchange(exchange, routingKey string, msg Message) error {
	msg.Timestamp = time.Now()

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mq.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// DeclareExchange declares an exchange
func (mq *MessageQueue) DeclareExchange(name, kind string) error {
	err := mq.channel.ExchangeDeclare(
		name,
		kind,
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", name, err)
	}
	return nil
}

// BindQueue binds a queue to an exchange
func (mq *MessageQueue) BindQueue(queue, exchange, routingKey string) error {
	err := mq.channel.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queue, exchange, err)
	}
	return nil
}
