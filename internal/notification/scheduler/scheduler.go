package scheduler

import (
	"context"
	"fmt"
	"github.com/ennwy/calendar/internal/app"
	noti "github.com/ennwy/calendar/internal/notification"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

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
	ctx     context.Context
	storage Storage
	period  noti.Period
	mq      *MessageQ
	mu      *sync.Mutex
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

func New(
	ctx context.Context, s Storage,
	log Logger, opts noti.MQProduce,
) (
	scheduler *Scheduler, err error,
) {
	l = log
	l.Info("[ + ] Scheduler created")

	scheduler = &Scheduler{
		ctx:     ctx,
		storage: s,
		period:  opts.P,
		mu:      &sync.Mutex{},
	}

	_ = scheduler.period.Set()

	scheduler.mq, err = NewQueue(opts)
	if err != nil {
		return nil, fmt.Errorf("new scheduler: %w", err)
	}

	if err = scheduler.storage.Connect(scheduler.ctx); err != nil {
		return nil, fmt.Errorf("new scheduler: storage conn: %w", err)
	}

	return scheduler, nil
}

func (s *Scheduler) Start() {
	l.Warn("Scheduler started")
	checkTimer := time.NewTicker(s.period.DBCheck)
	clearTimer := time.NewTicker(s.period.Clear)

	for {
		select {

		case <-checkTimer.C:
			if err := s.Publish(); err != nil {
				l.Error(err)
			}

		case <-clearTimer.C:
			if err := s.Clean(); err != nil {
				l.Error(err)
			}

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) Stop() (err error) {
	l.Warn("Scheduler Stopped")

	if err = s.mq.Close(); err != nil {
		return fmt.Errorf("scheduler stop: %w", err)
	}

	return nil
}

func (s *Scheduler) Publish() error {
	s.mu.Lock()
	events, err := s.storage.ListUpcoming(s.ctx, s.period.DBCheck)
	s.mu.Unlock()

	if err != nil {
		l.Error("scheduler: publish:", err)
		l.Error("events: ", events)
		return fmt.Errorf("scheduler: publish: list: %w", err)
	}

	if err = s.publishUpcoming(events); err != nil {
		l.Error("publish: publish event:", err)
		return fmt.Errorf("scheduler: publish: %w", err)
	}

	return nil
}

func (s *Scheduler) publishUpcoming(events *storage.Events) (err error) {
	l.Info("publish event: events found:", len(events.Events))

	for i, event := range events.Events {
		l.Info("[", i, "] publishing event:", event)

		err = s.publishEvent(event)
		if err != nil {
			l.Error("publish upcoming:", err)
			continue
		}
	}

	return nil
}

func (s *Scheduler) publishEvent(e storage.Event) (err error) {
	var bEvent []byte

	if bEvent, err = e.Marshall(); err != nil {
		return err
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

	return err
}

func (s *Scheduler) Clean() error {
	l.Info("scheduler: storage clean started")

	s.mu.Lock()
	err := s.storage.Clean(s.ctx, 365*storage.Day)
	s.mu.Unlock()

	if err != nil {
		return fmt.Errorf("start: clean: %w", err)
	}

	return nil
}
