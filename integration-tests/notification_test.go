package integration_tests

import (
	"context"
	"fmt"
	"github.com/ennwy/calendar/internal/logger"
	noti "github.com/ennwy/calendar/internal/notification"
	"github.com/ennwy/calendar/internal/notification/sender"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

type Config struct {
	Logger logger.Config  `yaml:"logger"`
	MQ     noti.MQConsume `yaml:"MQConsume"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Logger.Set()

	if err := config.MQ.Set(); err != nil {
		return nil, fmt.Errorf("sender: new configs: %w", err)
	}

	return config, nil
}

var l *logger.Logger

func TestNotification(t *testing.T) {
	c, err := NewConfig()
	require.NoError(t, err)

	l = logger.New(c.Logger.Level, c.Logger.OutputPath)
	ctx := context.Background()

	consumer, err := sender.New(ctx, l, c.MQ)
	require.NoError(t, err)

	user := createUser("Coban")

	event := storage.Event{
		Owner:  user,
		Title:  "Cool_Event_Title",
		Start:  processTime(time.Now().Add(61 * time.Minute)),
		Finish: processTime(time.Now().Add(300 * 24 * time.Hour)),
		Notify: 60,
	}
	err = createEventGRPC(&event)
	require.NoError(t, err)

	time.Sleep(20 * time.Second)
	log.Println("slept")

	messageCh, err := consumer.Start()
	require.NoError(t, err)

	timer := time.NewTimer(62 * time.Second)
	var e storage.Event

	for {
		select {

		case <-timer.C:
			require.Fail(t, "expected event")

		case message := <-messageCh:
			log.Println("in message case")
			err = e.Unmarshall(message.Body)
			require.NoError(t, err)
			require.Equal(t, event, e)

			err = message.Ack(false)
			require.NoError(t, err)
		}
		break
	}
}
