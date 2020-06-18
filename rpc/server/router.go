package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/rpc/proto"
)

func (s *Server) setRouter() {
	r := s.engine
	r.GET("/pipeline/list", s.listPipelines)
	r.POST("/pipeline/add", s.addPipeline)
	r.GET("/pipeline/ctrl", s.ctrlPipeline)
	r.GET("/pipeline/components", s.listPipelineComponents)
	r.GET("/pipeline/processors", s.listPipelineProcessors)
	r.GET("/pipeline/visualize", s.pipelineVisualize)
	r.GET("/pipeline/vars", s.pipelineVars)
	r.GET("/pipeline/config", s.pipelineConfig)
	r.GET("/pipeline/generate-config", s.generateConfig)

	r.GET("/component/list", s.listComponents)
	r.GET("/component/config", s.componentConfig)

	r.GET("/processor/list", s.listProcessors)
	r.GET("/processor/config", s.processorConfig)

	r.GET("/plugin/list", s.listPlugins)
	r.POST("/plugin/upload", s.uploadPlugin)
	r.POST("/plugin/open", s.openPlugin)

	r.GET("/metadata", func(c *gin.Context) {
		Success(c, proto.MetadataView{
			PluginPaths:         s.metadata.ListPaths(FileTypePlugin),
			PipelineConfigPaths: s.metadata.ListPaths(FileTypePipelineConfig),
			HTTPAddr:            s.options.HTTPAddr,
		})
	})
}
