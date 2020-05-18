package pipeline

type Config struct {
	Name       string
	Components map[string]string // key: name, value: rawConfig
	Stream     *StreamConfig     // key: name, value: FlowConfig
}

func (c *Config) AddStream(conf StreamConfig) *StreamConfig {
	if c.Stream == nil {
		c.Stream = &conf
	} else {
		c.Stream.AddStream(conf)
	}

	return c.Stream
}
