package server

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/processor"
	"github.com/shima-park/nezha/rpc/proto"
)

func (s *Server) listProcessors(c *gin.Context) {
	var res []proto.ProcessorView
	for _, c := range processor.ListFactory() {
		res = append(res, proto.ProcessorView{
			Name:         c.Name,
			Description:  c.Factory.Description(),
			SampleConfig: c.Factory.SampleConfig(),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	Success(c, res)
}

func (s *Server) processorConfig(c *gin.Context) {
	factory, err := processor.GetFactory(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	c.String(http.StatusOK, factory.SampleConfig())
}
