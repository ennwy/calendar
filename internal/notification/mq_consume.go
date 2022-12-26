package notification

import (
	"fmt"
	"os"
	"strconv"
)

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
			return nil, fmt.Errorf("configs parsing: %s: %w", arg, err)
		}

		vars[arg] = v
	}

	return vars, nil
}
