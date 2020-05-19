package pipeline

type StreamConfig struct {
	Name            string         `yaml:"name"`
	ProcessorName   string         `yaml:"processor_name"`
	ProcessorConfig string         `yaml:"processor_config"`
	Childs          []StreamConfig `yaml:"childs"`
}

func (sc *StreamConfig) AddStream(conf StreamConfig) *StreamConfig {
	sc.Childs = append(sc.Childs, conf)
	return sc
}
