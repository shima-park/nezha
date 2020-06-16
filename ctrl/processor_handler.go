package ctrl

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/processor"
)

type ProcessorView struct {
	Name         string `json:"name"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
}

func (ctrl *Ctrl) listProcessors(c *gin.Context) {
	var res []ProcessorView
	for _, c := range processor.ListFactory() {
		res = append(res, ProcessorView{
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

func (ctrl *Ctrl) processorConfig(c *gin.Context) {
	factory, err := processor.GetFactory(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	c.String(http.StatusOK, factory.SampleConfig())
}
