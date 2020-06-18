package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/common/plugin"
	_ "github.com/shima-park/nezha/component/include"
	"github.com/shima-park/nezha/pipeline"
)

type Server struct {
	options         Options
	metadata        Metadata
	engine          *gin.Engine
	pipelineManager pipeline.PipelinerManager
}

func New(opts ...Option) (*Server, error) {
	c := &Server{
		options:         defaultOptions,
		engine:          gin.Default(),
		pipelineManager: pipeline.NewPipelinerManager(),
	}

	for _, opt := range opts {
		opt(&c.options)
	}

	return c, c.init()
}

func (c *Server) init() error {
	c.setRouter()

	var err error
	c.metadata, err = NewMetadata(c.options.MetadataPath)
	if err != nil {
		return err
	}

	for _, path := range c.metadata.ListPaths(FileTypePlugin) {
		err := plugin.LoadPlugins(path)
		if err != nil {
			return err
		}
	}

	for _, path := range c.metadata.ListPaths(FileTypePipelineConfig) {
		err := pipeline.LoadPipelineFromFile(path, c.pipelineManager)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Server) Serve() error {
	for _, p := range c.pipelineManager.List() {
		if p.GetConfig().Bootstrap {
			if err := p.Start(); err != nil {
				return err
			}
		}
	}

	if c.options.HTTPAddr != "" {
		return c.engine.Run(c.options.HTTPAddr)
	}
	return nil
}

func (c *Server) Stop() {
	for _, p := range c.pipelineManager.List() {
		p.Stop()
	}
}