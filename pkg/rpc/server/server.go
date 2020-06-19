package server

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/common/log"
	"github.com/shima-park/lotus/common/plugin"
	"github.com/shima-park/lotus/pipeline"
	"gopkg.in/yaml.v2"
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
		err := loadPipelineFromFile(path, c.pipelineManager)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadPipelineFromFile(path string, pm pipeline.PipelinerManager) error {
	log.Info("loading pipeline from: %s", path)

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var conf pipeline.Config
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return err
	}

	_, err = pm.AddPipeline(conf)
	if err != nil {
		return err
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
