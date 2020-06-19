package server

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/component"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

func (s *Server) listComponents(c *gin.Context) {
	var res []proto.ComponentView
	for _, c := range component.ListFactory() {
		res = append(res, proto.ComponentView{
			Name:         c.Name,
			SampleConfig: c.Factory.SampleConfig(),
			Description:  c.Factory.Description(),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	Success(c, res)
}

func (s *Server) componentConfig(c *gin.Context) {
	factory, err := component.GetFactory(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, factory.SampleConfig())
}
