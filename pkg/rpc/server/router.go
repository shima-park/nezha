package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

func (s *Server) setRouter() {
	r := s.engine
	r.GET("/pipeline/generate-config", s.generateConfig)
	r.POST("/pipeline/add", s.addPipeline)
	r.POST("/pipeline/recreate", s.recreatePipeline)
	r.GET("/pipeline/ctrl", s.ctrlPipeline)
	r.GET("/pipeline/list", s.listPipelines)
	r.GET("/pipeline", s.findPipeline)

	r.GET("/component/list", s.listComponents)
	r.GET("/component", s.findComponent)

	r.GET("/processor/list", s.listProcessors)
	r.GET("/processor", s.findProcessor)

	r.GET("/plugin/list", s.listPlugins)
	r.POST("/plugin/upload", s.uploadPlugin)
	r.POST("/plugin/open", s.openPlugin)

	r.GET("/metadata", func(c *gin.Context) {
		Success(c, proto.MetadataView{
			PluginPaths:         s.metadata.ListPaths(proto.FileTypePlugin),
			PipelineConfigPaths: s.metadata.ListPaths(proto.FileTypePipelineConfig),
			HTTPAddr:            s.options.HTTPAddr,
		})
	})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, proto.Result{
		Data: data,
	})
}

func Failed(c *gin.Context, err error) {
	c.JSON(http.StatusOK, proto.Result{
		Code: http.StatusInternalServerError,
		Msg:  err.Error(),
	})
}
