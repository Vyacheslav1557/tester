package pkg

import amqp "github.com/rabbitmq/amqp091-go"

func NewRabbitClient(dsn string) (*amqp.Connection, error) {
	return amqp.Dial(dsn)
}
