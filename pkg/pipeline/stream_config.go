package pipeline

type StreamConfig struct {
	Name       string         `yaml:"name"`
	Childs     []StreamConfig `yaml:"childs"`
	Replica    int            `yaml:"replica"`
	BufferSize int            `yaml:"buffer_size"`
}
