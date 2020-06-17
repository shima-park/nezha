package ctrl

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/common/plugin"
)

func (ctrl *Ctrl) listPlugins(c *gin.Context) {
	var res []interface{}
	for _, p := range plugin.List() {
		res = append(res, struct {
			Path     string `json:"path"`
			Module   string `json:"module"`
			OpenTime string `json:"open_time"`
		}{
			Path:     p.Path,
			Module:   p.Module,
			OpenTime: fmt.Sprint(p.OpenTime),
		})
	}

	Success(c, res)
}

func (ctrl *Ctrl) openPlugin(c *gin.Context) {
	var req struct {
		Path string
	}
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
