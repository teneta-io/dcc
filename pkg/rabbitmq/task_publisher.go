package rabbitmq

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	QueueName = "tasks"
)

type TaskPublisher struct {
	client    *RabbitMQ
	queueName string
	ch        *Channel
	queue     chan *Message
}

func NewTaskPublisher(client *RabbitMQ) (*TaskPublisher, error) {
	publisher := &TaskPublisher{
		client:    client,
		queueName: QueueName,
		queue:     make(chan *Message, 1000),
	}

	ch, err := publisher.client.connection.Channel()

	if err != nil {
		return nil, err
	}

	publisher.ch = ch

	return publisher, nil
}

func (publisher *TaskPublisher) Publish(body []byte) {
	if err := publisher.publish(&Message{
		body: body,
	}); err != nil {
		zap.S().Error(err.Error())
	}
}

func (publisher *TaskPublisher) publish(message *Message) error {
	if publisher.client.connection == nil {
		zap.S().Panic(ErrSendBeforeEstablishConnection.Error())
	}

	queue, err := publisher.ch.QueueDeclare(
		publisher.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		zap.S().Error(err.Error())
		return err
	}

	err = publisher.ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			Body: message.body,
		},
	)

	return err
}

func (publisher *TaskPublisher) Disconnect() error {
	return publisher.ch.Close()
}
