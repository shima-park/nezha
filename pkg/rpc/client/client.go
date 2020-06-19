package client

import (
	pipe "github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

type Pipeline interface {
	List() ([]proto.PipelineView, error)
	Config(name string) (string, error)
	GenerateConfig(name string, components, processors []string) (string, error)
	Add(conf pipe.Config) error
	Control(cmd proto.ControlCommand, names []string) error
	ListComponents(name string) ([]proto.ComponentView, error)
	ListProcessors(name string) ([]proto.ProcessorView, error)
	Visualize(name string, format proto.VisualizeFormat) error
	Vars(name string) (map[string]string, error)
}

type Component interface {
	List() ([]proto.ComponentView, error)
	Config(name string) (string, error)
}

type Processor interface {
	List() ([]proto.ProcessorView, error)
	Config(name string) (string, error)
}

type Plugin interface {
	List() ([]proto.PluginView, error)
	Open(path string) error
	Upload(path string) error
}

type Server interface {
	Metadata() (proto.MetadataView, error)
}

type Client struct {
	Pipeline
	Component
	Processor
	Plugin
	Server

	addr string
}

func NewClient(addr string) *Client {
	addr = normalizeURL(addr)
	b := apiBuilder{addr}
	return &Client{
		Pipeline:  &pipeline{b},
		Component: &component{b},
		Processor: &processor{b},
		Plugin:    &plugin{b},
		Server:    &server{b},
	}
}

type apiBuilder struct {
	addr string
}

func (b *apiBuilder) api(path string) string {
	return b.addr + path
}
