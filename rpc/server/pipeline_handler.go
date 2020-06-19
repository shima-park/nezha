package server

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
	"github.com/shima-park/lotus/component"
	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/lotus/processor"
	"github.com/shima-park/nezha/rpc/proto"
	"gopkg.in/yaml.v2"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, proto.Result{
		Data: data,
	})
}

func Failed(c *gin.Context, err error) {
	c.JSON(http.StatusOK, proto.Result{
		Code: http.StatusInternalServerError,
		Msg:  err.Error(),
	})
}

func (s *Server) handlePipelineByName(c *gin.Context, callback func(c *gin.Context, pipe pipeline.Pipeliner)) {
	p := s.pipelineManager.Find(c.Query("name"))
	if p == nil {
		c.Status(http.StatusNotFound)
		return
	}

	callback(c, p)
}

func (s *Server) listPipelines(c *gin.Context) {
	var res []proto.PipelineView
	for _, p := range s.pipelineManager.List() {
		res = append(res, proto.PipelineView{
			Name:          p.Name(),
			State:         p.State().String(),
			Schedule:      p.GetConfig().Schedule,
			Bootstrap:     p.GetConfig().Bootstrap,
			StartTime:     p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_START_TIME).String(),
			ExitTime:      p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_EXIT_TIME).String(),
			RunTimes:      p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_RUN_TIMES).String(),
			NextRunTime:   p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_NEXT_RUN_TIME).String(),
			LastStartTime: p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_LAST_START_TIME).String(),
			LastEndTime:   p.Monitor().Get(pipeline.METRICS_KEY_PIPELINE_LAST_END_TIME).String(),
		})
	}
	Success(c, res)
}

func (s *Server) addPipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}

	name := conf.Name + ".yaml"
	path := s.metadata.GetPath(FileTypePipelineConfig, name)
	if s.metadata.ExistsPath(FileTypePipelineConfig, path) {
		Failed(c, fmt.Errorf("The pipeline name(%s) is exists", name))
		return
	}

	err := os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		Failed(c, err)
		return
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		Failed(c, err)
		return
	}

	err = ioutil.WriteFile(path, data, 0640)
	if err != nil {
		Failed(c, err)
		return
	}

	_, err = s.pipelineManager.AddPipeline(conf)
	if err != nil {
		Failed(c, err)
		return
	}

	err = s.metadata.AddPath(FileTypePipelineConfig, path)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) ctrlPipeline(c *gin.Context) {
	methodMap := map[string]func(names ...string) error{
		"start":   s.pipelineManager.Start,
		"stop":    s.pipelineManager.Stop,
		"restart": s.pipelineManager.Restart,
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

func (s *Server) listPipelineComponents(c *gin.Context) {
	s.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		var res []proto.ComponentView
		for _, c := range pipe.ListComponents() {
			res = append(res, proto.ComponentView{
				Name:         c.Name,
				RawConfig:    c.RawConfig,
				SampleConfig: c.Factory.SampleConfig(),
				Description:  c.Factory.Description(),
				InjectName:   c.Component.Instance().Name(),
				ReflectType:  c.Component.Instance().Type().String(),
				ReflectValue: c.Component.Instance().Value().String(),
			})
		}

		Success(c, res)
	})
}

func (s *Server) listPipelineProcessors(c *gin.Context) {
	s.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		var res []proto.ProcessorView
		for _, c := range pipe.ListProcessors() {
			res = append(res, proto.ProcessorView{
				Name:         c.Name,
				RawConfig:    c.RawConfig,
				Description:  c.Factory.Description(),
				SampleConfig: c.Factory.SampleConfig(),
			})
		}

		c.JSON(http.StatusOK, res)
	})
}

func (s *Server) pipelineVisualize(c *gin.Context) {
	s.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		err := pipe.Visualize(c.Writer, c.Query("format"))
		if err != nil {
			Failed(c, err)
			return
		}
	})
}

func (s *Server) generateConfig(c *gin.Context) {
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

	b, err := yaml.Marshal(conf)
	if err != nil {
		Failed(c, err)
		return
	}
	c.String(200, string(b))
}

func (s *Server) pipelineVars(c *gin.Context) {
	s.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		res := map[string]string{}
		pipe.Monitor().Do(func(namespace string, kv expvar.KeyValue) {
			res[namespace+kv.Key] = kv.Value.String()
		})
		Success(c, res)
	})
}

func (s *Server) pipelineConfig(c *gin.Context) {
	s.handlePipelineByName(c, func(c *gin.Context, pipe pipeline.Pipeliner) {
		config, err := yaml.Marshal(pipe.GetConfig())
		if err != nil {
			Failed(c, err)
			return
		}

		Success(c, string(config))
	})
}
