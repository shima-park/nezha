package server

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

func (s *Server) listPipelines(c *gin.Context) {
	res, err := s.Pipeline.List()
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) addPipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}

	err := s.Pipeline.Add(conf)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) ctrlPipeline(c *gin.Context) {
	err := s.Pipeline.Control(proto.ControlCommand(c.Query("cmd")), c.QueryArray("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) generateConfig(c *gin.Context) {
	name := c.Query("name")
	schedule := c.Query("schedule")
	components := strings.Split(c.Query("components"), ",")
	processors := strings.Split(c.Query("processors"), ",")
	config, err := s.Pipeline.GenerateConfig(name, schedule, components, processors)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, config)
}

func (s *Server) findPipeline(c *gin.Context) {
	pipe, err := s.Pipeline.Find(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, pipe)
}

func (s *Server) recreatePipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}
	err := s.Pipeline.Recreate(conf)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}
