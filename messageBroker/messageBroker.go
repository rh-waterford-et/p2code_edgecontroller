package messagebroker

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageBroker struct {
	amqpConnection *amqp.Connection
	channel        *amqp.Channel
	// Q list of channels ??
}

func RegisterBroker(user, password, host, port string) (*MessageBroker, error) {
	amqpConn := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	conn, err := amqp.Dial(amqpConn)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	msgBroker := MessageBroker{}
	msgBroker.amqpConnection = conn
	msgBroker.channel = ch
	return &msgBroker, nil
}

// Set up channel and close once done - shouldnt have channel in broker obj
func (b *MessageBroker) SendSwitchCommand(body string) error {
	switchQueue, err := b.channel.QueueDeclare(
		"switchImage", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = b.channel.PublishWithContext(ctx,
		"",               // exchange
		switchQueue.Name, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/x-sh",
			Body:        []byte(body),
		})

	return err
}

func (b *MessageBroker) Read() {

}

// Q is this right
func (b *MessageBroker) Close() {
	b.channel.Close()
	b.amqpConnection.Close()
}
