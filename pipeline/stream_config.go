package pipeline

type StreamConfig struct {
	Name       string         `yaml:"name"`
	Childs     []StreamConfig `yaml:"childs,omitempty"`
	Replica    int            `yaml:"replica,omitempty"`
	BufferSize int            `yaml:"buffer_size,omitempty"`
}
