package proto

import (
	pipe "github.com/shima-park/lotus/pipeline"
)

type Pipeline interface {
	GenerateConfig(name, schdule string, components, processors []string) (*pipe.Config, error)
	Add(conf pipe.Config) error
	Recreate(conf pipe.Config) error
	List() ([]PipelineView, error)
	Find(name string) (*PipelineView, error)
	Control(cmd ControlCommand, names []string) error
}

type Component interface {
	List() ([]ComponentView, error)
	Find(name string) (*ComponentView, error)
}

type Processor interface {
	List() ([]ProcessorView, error)
	Find(name string) (*ProcessorView, error)
}

type Plugin interface {
	List() ([]PluginView, error)
	Open(path string) error
	Add(path string) error
}

type Server interface {
	Metadata() (MetadataView, error)
}

type FileType string

const (
	FileTypePlugin         FileType = "plugins"
	FileTypePipelineConfig FileType = "pipelines"
)

type Metadata interface {
	AddPath(ft FileType, path string) error
	RemovePath(ft FileType, path string) error
	GetPath(ft FileType, filename string) string
	ExistsPath(ft FileType, path string) bool
	ListPaths(ft FileType) []string
	Overwrite(ft FileType, path string, data []byte) error
}
