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

type Scheduler struct {
	ctx     context.Context
	q       amqp.Queue
	conn    *amqp.Connection
	ch      *amqp.Channel
	storage Storage
	opts    noti.MQPublish
}

var l Logger

func New(ctx context.Context, storage Storage, log Logger, opts noti.MQProduce) (s *Scheduler, err error) {
	l = log
	l.Info("Scheduler created")

	s = &Scheduler{
		ctx:     ctx,
		storage: storage,
		opts:    opts.Publish,
	}

	if s.conn, err = amqp.Dial(opts.Q.URL); err != nil {
		return nil, fmt.Errorf("scheduler start: %w", err)
	}

	if s.ch, err = s.conn.Channel(); err != nil {
		return nil, fmt.Errorf("start channel: %w", err)
	}

	if err = s.ch.ExchangeDeclare(
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

	s.q, err = s.ch.QueueDeclare(
		opts.Q.Name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		nil,
	)

	return s, err
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

	if err = s.ch.Close(); err != nil {
		return fmt.Errorf("scheduler stop: chan close: %w", err)
	}

	return s.conn.Close()
}

func (s *Scheduler) publish() {
	var events []storage.Event
	var err error

	for t := time.NewTicker(period); ; {
		select {
		case <-t.C:
			events, err = s.storage.ListUpcoming(s.ctx, period)

			if err != nil {
				l.Error("scheduler: publish:", err)
				l.Error("events: ", events)
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

func (s *Scheduler) publishEvent(events []storage.Event) (err error) {
	l.Info("publish event: events found:", len(events))
	var bEvent []byte

	for i, event := range events {
		l.Info("[", i, "] publishing event:", event)

		if bEvent, err = event.Marshall(); err != nil {
			l.Error("scheduler: publish: selecting: event:", err)
			continue
		}

		err = s.ch.Publish(
			"",
			s.q.Name,
			s.opts.Mandatory,
			s.opts.Immediate,
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
