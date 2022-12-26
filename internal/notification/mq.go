package notification

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
