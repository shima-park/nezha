package pipeline

type Config struct {
	Name       string              `yaml:"name"`
	Components []map[string]string `yaml:"components"` // key: name, value: rawConfig
	Stream     *StreamConfig       `yaml:"stream"`     // key: name, value: FlowConfig
}

func (c *Config) AddStream(conf StreamConfig) *StreamConfig {
	if c.Stream == nil {
		c.Stream = &conf
	} else {
		c.Stream.AddStream(conf)
	}

	return c.Stream
}
