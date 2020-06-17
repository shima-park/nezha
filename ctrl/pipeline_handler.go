package ctrl

import (
	"errors"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/component"
	"github.com/shima-park/nezha/pipeline"
	"github.com/shima-park/nezha/processor"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{
		Data: data,
	})
}

func Failed(c *gin.Context, err error) {
	c.JSON(http.StatusOK, Result{
		Code: http.StatusInternalServerError,
		Msg:  err.Error(),
	})
}

func (ctrl *Ctrl) handlePipelineByName(c *gin.Context, callback func(c *gin.Context, pipe pipeline.Pipeliner)) {
	p := ctrl.pipelineManager.Find(c.Query("name"))
	if p == nil {
		c.Status(http.StatusNotFound)
		return
	}

	callback(c, p)
}

func (ctrl *Ctrl) listPipelines(c *gin.Context) {
	var res []interface{}
	for _, p := range ctrl.pipelineManager.List() {
		res = append(res, struct {
			Name      string `json:"name"`
			State     string `json:"state"`
			Schedule  string `json:"schedule"`
			Bootstrap bool   `json:"bootstrap"`
			StartTime string `json:"start_time"`
			ExitTime  string `json:"exit_time"`
			RunTimes  string `json:"run_times"`
		}{
			Name:      p.Name(),
			State:     p.State().String(),
			Schedule:  p.GetConfig().Schedule,
			Bootstrap: p.GetConfig().Bootstrap,
			StartTime: p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_START_TIME).String(),
			ExitTime:  p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_EXIT_TIME).String(),
			RunTimes:  p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_RUN_TIMES).String(),
		})
	}
	Success(c, res)
}

func (ctrl *Ctrl) addPipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}

	name := conf.Name + ".yaml"
	path := ctrl.metadata.GetPath(FileTypePipelineConfig, name)
	if ctrl.metadata.ExistsPath(FileTypePipelineConfig, path) {
		Failed(c, fmt.Errorf("The pipeline name(%s) is exists", name))
		return
	}

	err := os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		Failed(c, err)
		return
	}

	data, err := conf.Marshal()
	if err != nil {
		Failed(c, err)
		return
	}

	err = ioutil.WriteFile(path, data, 0640)
	if err != nil {
		Failed(c, err)
		return
	}

	_, err = ctrl.pipelineManager.AddPipeline(conf)
	if err != nil {
		Failed(c, err)
		return
	}

	err = ctrl.metadata.AddPath(FileTypePipelineConfig, path)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (ctrl *Ctrl) ctrlPipeline(c *gin.Context) {
	methodMap := map[string]func(names ...string) error{
		"start":   ctrl.pipelineManager.Start,
		"stop":    ctrl.pipelineManager.Stop,
		"restart": ctrl.pipelineManager.Restart,
	}

	cmd := c.Query("cmd")
	m, ok := methodMap[cmd]
	if !ok {
		Failed(c, errors.New("Unsupported method "+cmd))
		return
	}

	err := m(c.QueryArray("name")...)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (ctrl *Ctrl) listPipelineComponents(c *gin.Context) {
	ctrl.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		var res []interface{}
		for _, c := range pipe.ListComponents() {
			res = append(res, struct {
				Name         string `json:"name"`
				RawConfig    string `json:"raw_config"`
				SampleConfig string `json:"sample_config"`
				Description  string `json:"description"`
				InjectName   string `json:"inject_name"`
				ReflectType  string `json:"reflect_type"`
				ReflectValue string `json:"reflect_value"`
			}{
				c.Name,
				c.RawConfig,
				c.Factory.SampleConfig(),
				c.Factory.Description(),
				c.Component.Instance().Name(),
				c.Component.Instance().Type().String(),
				c.Component.Instance().Value().String(),
			})
		}

		Success(c, res)
	})
}

func (ctrl *Ctrl) listPipelineProcessors(c *gin.Context) {
	ctrl.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		var res []interface{}
		for _, c := range pipe.ListProcessors() {
			res = append(res, struct {
				Name      string `json:"name"`
				RawConfig string `json:"raw_config"`
			}{
				c.Name,
				c.RawConfig,
			})
		}

		c.JSON(http.StatusOK, res)
	})
}

func (ctrl *Ctrl) pipelineVisualize(c *gin.Context) {
	ctrl.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		err := pipe.Visualize(c.Writer, c.Query("format"))
		if err != nil {
			Failed(c, err)
			return
		}
	})
}

func (ctrl *Ctrl) generateConfig(c *gin.Context) {
	name := c.Query("name")
	components := strings.Split(c.Query("components"), ",")
	processors := strings.Split(c.Query("processors"), ",")

	var componentConfigs []map[string]string
	for _, name := range components {
		name = strings.TrimSpace(name)
		f, err := component.GetFactory(name)
		if err != nil {
			Failed(c, err)
			return
		}
		componentConfigs = append(componentConfigs, map[string]string{
			name: f.SampleConfig(),
		})
	}

	var processorConfigs []map[string]string
	streamConfig := &pipeline.StreamConfig{}
	t := streamConfig
	for i, name := range processors {
		name = strings.TrimSpace(name)
		f, err := processor.GetFactory(name)
		if err != nil {
			Failed(c, err)
			return
		}

		t.Name = name
		if i != len(processors)-1 { // 防止加上最后一个空childs
			t.Childs = []pipeline.StreamConfig{
				pipeline.StreamConfig{},
			}
			t = &t.Childs[0]
		}

		processorConfigs = append(processorConfigs, map[string]string{
			name: f.SampleConfig(),
		})
	}

	conf := pipeline.Config{
		Name:       name,
		Schedule:   c.Query("schedule"),
		Components: componentConfigs,
		Processors: processorConfigs,
		Stream:     *streamConfig,
	}

	b, err := config.Marshal(conf)
	if err != nil {
		Failed(c, err)
		return
	}
	c.String(200, string(b))
}

func (ctrl *Ctrl) pipelineVars(c *gin.Context) {
	ctrl.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		name := c.Query("name")
		var res []map[string]string
		for _, p := range ctrl.pipelineManager.List() {
			if name != "" && name != p.Name() {
				continue
			}

			m := map[string]string{}

			p.Monitor().Do(func(namespace string, kv expvar.KeyValue) {
				m[namespace+kv.Key] = kv.Value.String()
			})
			res = append(res, m)
		}

		Success(c, res)
	})
}

func (ctrl *Ctrl) pipelineConfig(c *gin.Context) {
	ctrl.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		config, err := pipe.GetConfig().Marshal()
		if err != nil {
			Failed(c, err)
			return
		}

		c.String(http.StatusOK, string(config))
	})
}
