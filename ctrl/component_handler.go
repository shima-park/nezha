package ctrl

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/component"
)

type ComponentView struct {
	Name         string `json:"name"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
}

func (ctrl *Ctrl) listComponents(c *gin.Context) {
	var res []ComponentView
	for _, c := range component.ListFactory() {
		res = append(res, ComponentView{
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

func (ctrl *Ctrl) componentConfig(c *gin.Context) {
	factory, err := component.GetFactory(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	c.String(http.StatusOK, factory.SampleConfig())
}
