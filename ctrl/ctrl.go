package ctrl

import (
	"github.com/gin-gonic/gin"
	_ "github.com/shima-park/nezha/component/include"
	"github.com/shima-park/nezha/pipeline"
	"github.com/shima-park/nezha/plugin"
)

type Ctrl struct {
	options         Options
	engine          *gin.Engine
	pipelineManager pipeline.PipelinerManager
}

func New(opts ...Option) (*Ctrl, error) {
	c := &Ctrl{
		options:         Options{},
		engine:          gin.Default(),
		pipelineManager: pipeline.NewPipelinerManager(),
	}
	c.options.Apply(opts...)
	c.setRouter()
	return c, nil
}

func (c *Ctrl) Serve() error {
	if len(c.options.PluginPaths) > 0 {
		for _, path := range c.options.PluginPaths {
			err := plugin.LoadPlugins(path)
			if err != nil {
				return err
			}
		}
	}

	if len(c.options.PipelinePaths) > 0 {
		for _, path := range c.options.PipelinePaths {
			err := pipeline.LoadPipelineFromFile(path, c.pipelineManager)
			if err != nil {
				return err
			}
		}
	}

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

func (c *Ctrl) Stop() {
	for _, p := range c.pipelineManager.List() {
		p.Stop()
	}
}
