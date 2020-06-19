package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/common/plugin"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

func (s *Server) listPlugins(c *gin.Context) {
	var res []proto.PluginView
	for _, p := range plugin.List() {
		res = append(res, proto.PluginView{
			Path:     p.Path,
			Module:   p.Module,
			OpenTime: fmt.Sprint(p.OpenTime),
		})
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

	err = plugin.LoadPlugins(req.Path)
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
	path := s.metadata.GetPath(FileTypePlugin, filename)
	if s.metadata.ExistsPath(FileTypePlugin, path) {
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

	err = s.metadata.AddPath(FileTypePlugin, path)
	if err != nil {
		Failed(c, err)
		return
	}

	err = plugin.LoadPlugins(path)
	if err != nil {
		_ = s.metadata.RemovePath(FileTypePlugin, path)
		Failed(c, err)
		return
	}

	Success(c, nil)
}
