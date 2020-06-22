package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

func (s *Server) listPlugins(c *gin.Context) {
	res, err := s.Plugin.List()
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, res)
}

func (s *Server) openPlugin(c *gin.Context) {
	var req proto.PluginOpenRequest
	err := c.BindJSON(&req)
	if err != nil {
		Failed(c, err)
		return
	}

	err = s.Plugin.Open(req.Path)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) uploadPlugin(c *gin.Context) {
	pluginFile, err := c.FormFile("plugin")
	if err != nil {
		Failed(c, err)
		return
	}

	filename := filepath.Base(pluginFile.Filename)
	path := s.metadata.GetPath(proto.FileTypePlugin, filename)
	if s.metadata.ExistsPath(proto.FileTypePlugin, path) {
		Failed(c, fmt.Errorf("The plugin name(%s) is exists", path))
		return
	}

	err = os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		Failed(c, err)
		return
	}

	if err := c.SaveUploadedFile(pluginFile, path); err != nil {
		Failed(c, err)
		return
	}

	err = s.Plugin.Add(path)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}
