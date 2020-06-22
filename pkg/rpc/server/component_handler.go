package server

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) listComponents(c *gin.Context) {
	res, err := s.Component.List()
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, res)
}

func (s *Server) findComponent(c *gin.Context) {
	comp, err := s.Component.Find(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, comp)
}
