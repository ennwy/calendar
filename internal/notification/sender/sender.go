package sender

import (
	"context"
	"fmt"
	"github.com/ennwy/calendar/internal/app"
	noti "github.com/ennwy/calendar/internal/notification"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/streadway/amqp"
)

type Logger interface {
	app.Logger
}

type Sender struct {
	ctx  context.Context
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
	opts noti.MQConsume
}

var l Logger

func NewSender(ctx context.Context, log Logger, opts noti.MQConsume) (s *Sender, err error) {
	l = log

	s = &Sender{ctx: ctx, opts: opts}
	if s.conn, err = amqp.Dial(opts.Q.URL); err != nil {
		return nil, fmt.Errorf("new sender dial: %w", err)
	}

	if s.ch, err = s.conn.Channel(); err != nil {
		return nil, fmt.Errorf("new sender: ch conn: %w", err)
	}
	l.Warn("sender created")
	return s, nil
}

func (s *Sender) Start() (<-chan amqp.Delivery, error) {
	messageCh, err := s.ch.Consume(
		s.opts.Q.Name,
		s.opts.Consumer,
		s.opts.AutoAck,
		s.opts.Exclusive,
		s.opts.NoLocal,
		s.opts.NoWait,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("consumer create: %w", err)
	}

	l.Info("listening started")
	return messageCh, nil
}

func (s *Sender) PrintMessages(messageCh <-chan amqp.Delivery) {
	var e storage.Event
	var counter int

	for message := range messageCh {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		if err := e.Unmarshall(message.Body); err != nil {
			l.Error("unmarshall event:", err)
			continue
		}

		l.Info("[", counter, "]", e)
		counter++

		if err := message.Ack(false); err != nil {
			l.Error("message ack:", err)
		}
	}
}

func (s *Sender) Stop() (err error) {
	if err = s.ch.Close(); err != nil {
		return fmt.Errorf("sender close: %w", err)
	}

	if err = s.conn.Close(); err != nil {
		return fmt.Errorf("sender close: %w", err)
	}

	return nil
}
