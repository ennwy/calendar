package notification

import (
	"fmt"
	"os"
	"time"
)

type MQPublish struct {
	Mandatory bool `yaml:"mandatory"`
	Immediate bool `yaml:"immediate"`
}

type MQProduce struct {
	Q          QConfig   `yaml:"q"`
	Publish    MQPublish `yaml:"publish"`
	P          Period
	Durable    bool `yaml:"durable"`
	AutoDelete bool `yaml:"autoDelete"`
	Exclusive  bool `yaml:"exclusive"`
	NoWait     bool `yaml:"noWait"`
}

func (q *MQProduce) Set() error {
	q.Q.Name = os.Getenv("MQ_PRODUCE_NAME")
	q.Q.URL = os.Getenv("MQ_URL")
	v, err := ParseBool(
		mqImmediate,
		mqMandatory,
		mqDurable,
		mqAutoDelete,
		mqExclusive,
		mqNoWait,
	)

	if err != nil {
		return fmt.Errorf("notification: %w", err)
	}

	q.Publish.Immediate = v[mqImmediate]
	q.Publish.Mandatory = v[mqMandatory]
	q.Durable = v[mqDurable]
	q.AutoDelete = v[mqAutoDelete]
	q.Exclusive = v[mqExclusive]
	q.NoWait = v[mqNoWait]

	q.P.DBCheck = 1 * time.Minute
	q.P.Clear = 24 * time.Hour

	_ = q.P.Set()

	return nil
}

type Period struct {
	DBCheck time.Duration
	Clear   time.Duration
}

func (p *Period) Set() (err error) {
	s := os.Getenv("PERIOD_CLEAR")
	if p.Clear, err = time.ParseDuration(s); err != nil {
		return fmt.Errorf("PERIOD_CLEAR var is not correct")
	}

	return nil
}
