package pipeline

type StreamConfig struct {
	Name            string
	ProcessorName   string
	ProcessorConfig string
	Childs          []StreamConfig
}

func (sc *StreamConfig) AddStream(conf StreamConfig) *StreamConfig {
	sc.Childs = append(sc.Childs, conf)
	return sc
}
