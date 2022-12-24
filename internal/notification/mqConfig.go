package notification

import (
	"fmt"
	"os"
	"strconv"
)

const (
	mqMandatory = "MQ_MANDATORY"
	mqImmediate = "MQ_IMMEDIATE"

	mqDurable    = "MQ_DURABLE"
	mqAutoDelete = "MQ_AUTO_DELETE"
	mqExclusive  = "MQ_EXCLUSIVE"
	mqNoWait     = "MQ_NO_WAIT"
	mqAutoAck    = "MQ_AUTO_ACK"
	mqNoLocal    = "MQ_NO_LOCAL"
)

type QConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type MQPublish struct {
	Mandatory bool `yaml:"mandatory"`
	Immediate bool `yaml:"immediate"`
}

type MQProduce struct {
	Q          QConfig   `yaml:"q"`
	Publish    MQPublish `yaml:"publish"`
	Durable    bool      `yaml:"durable"`
	AutoDelete bool      `yaml:"autoDelete"`
	Exclusive  bool      `yaml:"exclusive"`
	NoWait     bool      `yaml:"noWait"`
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

	return nil
}

type MQConsume struct {
	Q         QConfig `yaml:"q"`
	Consumer  string  `yaml:"consumer"`
	AutoAck   bool    `yaml:"autoAck"`
	Exclusive bool    `yaml:"exclusive"`
	NoLocal   bool    `yaml:"noLocal"`
	NoWait    bool    `yaml:"noWait"`
}

func (q *MQConsume) Set() error {
	q.Q.Name = os.Getenv("MQ_CONSUME_NAME")
	q.Q.URL = os.Getenv("MQ_URL")
	q.Consumer = os.Getenv("MQ_CONSUMER")

	v, err := ParseBool(
		mqAutoAck,
		mqExclusive,
		mqNoLocal,
		mqNoWait,
	)
	if err != nil {
		return fmt.Errorf("notification: %w", err)
	}

	q.AutoAck = v[mqAutoAck]
	q.Exclusive = v[mqExclusive]
	q.NoLocal = v[mqNoLocal]
	q.NoWait = v[mqNoWait]

	return nil
}

func ParseBool(args ...string) (map[string]bool, error) {
	vars := make(map[string]bool, len(args))

	for _, arg := range args {
		v, err := strconv.ParseBool(os.Getenv(arg))

		if err != nil {
			return nil, fmt.Errorf("config parsing: %s: %w", arg, err)
		}

		vars[arg] = v
	}

	return vars, nil
}
