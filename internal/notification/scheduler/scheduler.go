package scheduler

import (
	"context"
	"fmt"
	"github.com/ennwy/calendar/internal/app"
	noti "github.com/ennwy/calendar/internal/notification"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/streadway/amqp"
	"time"
)

const period = time.Minute

type Storage interface {
	app.CleanListener
}

type Logger interface {
	app.Logger
}

type MessageQ struct {
	Q    amqp.Queue
	Conn *amqp.Connection
	Ch   *amqp.Channel
	Opts noti.MQPublish
}

func (m *MessageQ) Close() (err error) {
	l.Warn("Queue Closed")

	if err = m.Ch.Close(); err != nil {
		return fmt.Errorf("queue close: chan close: %w", err)
	}

	if err = m.Conn.Close(); err != nil {
		return fmt.Errorf("queue close: Conn close: %w", err)
	}

	return nil
}

type Scheduler struct {
	mq  *MessageQ
	ctx context.Context
	//Q       amqp.Queue
	//Conn    *amqp.Connection
	//Ch      *amqp.Channel

	storage Storage
}

var l Logger

func NewQueue(opts noti.MQProduce) (s *MessageQ, err error) {
	s = &MessageQ{Opts: opts.Publish}

	if s.Conn, err = amqp.Dial(opts.Q.URL); err != nil {
		return nil, fmt.Errorf("new conn: %w", err)
	}

	if s.Ch, err = s.Conn.Channel(); err != nil {
		return nil, fmt.Errorf("start channel: %w", err)
	}

	if err = s.Ch.ExchangeDeclare(
		opts.Q.Name,
		amqp.ExchangeDirect,
		opts.Durable,
		opts.AutoDelete,
		false,
		opts.NoWait,
		nil,
	); err != nil {
		return nil, fmt.Errorf("exchange declare: %w", err)
	}

	s.Q, err = s.Ch.QueueDeclare(
		opts.Q.Name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		nil,
	)

	return s, err
}

func New(ctx context.Context, storage Storage, log Logger, opts noti.MQProduce) (s *Scheduler, err error) {
	l = log

	s = &Scheduler{
		ctx:     ctx,
		storage: storage,
	}

	s.mq, err = NewQueue(opts)
	if err != nil {
		return nil, fmt.Errorf("new scheduler: %w", err)
	}

	l.Info("Scheduler created")
	return s, nil
}

func (s *Scheduler) Start() error {
	l.Warn("Scheduler started")
	if err := s.storage.Connect(s.ctx); err != nil {
		return err
	}

	go func() {
		t := time.NewTicker(24 * time.Hour)

		for {
			select {

			case <-s.ctx.Done():
				return

			case <-t.C:
				if err := s.storage.Clean(s.ctx, 365*storage.Day); err != nil {
					l.Error("start: clean:", err)
				}
			}
		}
	}()

	s.publish()
	return nil
}

func (s *Scheduler) Stop() (err error) {
	l.Warn("Scheduler Stopped")

	if err = s.mq.Close(); err != nil {
		return fmt.Errorf("scheduler stop: %w", err)
	}

	return nil
}

func (s *Scheduler) publish() {
	var events *storage.Events
	var err error

	for t := time.NewTicker(period); ; {
		select {
		case <-t.C:
			events, err = s.storage.ListUpcoming(s.ctx, period)

			if err != nil {
				l.Error("scheduler: publish:", err)
				l.Error("events: ", events.Events)
				continue
			}

			if err = s.publishEvent(events); err != nil {
				l.Error("publish: publish event:", err)
			}

		case <-s.ctx.Done():
			l.Info("publish: ctx: done!")
			return
		}
	}
}

func (s *Scheduler) publishEvent(events *storage.Events) (err error) {
	l.Info("publish event: events found:", len(events.Events))
	var bEvent []byte

	for i, event := range events.Events {
		l.Info("[", i, "] publishing event:", event)

		if bEvent, err = event.Marshall(); err != nil {
			l.Error("scheduler: publish: selecting: event:", err)
			continue
		}

		err = s.mq.Ch.Publish(
			"",
			s.mq.Q.Name,
			s.mq.Opts.Mandatory,
			s.mq.Opts.Immediate,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        bEvent,
			},
		)

		if err != nil {
			l.Error("ERR publishing:", err)
		}
	}

	return nil
}
