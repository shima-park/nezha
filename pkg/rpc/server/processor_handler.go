package server

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) listProcessors(c *gin.Context) {
	res, err := s.Processor.List()
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, res)
}

func (s *Server) findProcessor(c *gin.Context) {
	proc, err := s.Processor.Find(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, proc)
}
