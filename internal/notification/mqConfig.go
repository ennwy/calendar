package notification

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

type MQConsume struct {
	Q         QConfig `yaml:"q"`
	Consumer  string  `yaml:"consumer"`
	AutoAck   bool    `yaml:"autoAck"`
	Exclusive bool    `yaml:"exclusive"`
	NoLocal   bool    `yaml:"noLocal"`
	NoWait    bool    `yaml:"noWait"`
}
