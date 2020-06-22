package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/shima-park/lotus/component"
	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/lotus/processor"
	"github.com/shima-park/nezha/pkg/rpc/proto"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigSuffix = ".yaml"
)

type pipelineService struct {
	metadata        proto.Metadata
	pipelineManager pipeline.PipelinerManager
}

func NewPipelineService(metadata proto.Metadata,
	pipelineManager pipeline.PipelinerManager) proto.Pipeline {
	return &pipelineService{
		metadata:        metadata,
		pipelineManager: pipelineManager,
	}
}

func (s *pipelineService) GenerateConfig(name, schedule string, components, processors []string) (*pipeline.Config, error) {
	var componentConfigs []map[string]string
	for _, name := range components {
		name = strings.TrimSpace(name)
		f, err := component.GetFactory(name)
		if err != nil {
			return nil, err
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
			return nil, err
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

	conf := &pipeline.Config{
		Name:       name,
		Schedule:   schedule,
		Components: componentConfigs,
		Processors: processorConfigs,
		Stream:     *streamConfig,
	}

	return conf, nil
}

func (s *pipelineService) Add(conf pipeline.Config) error {
	path := s.getConfigPath(conf.Name)
	if s.metadata.ExistsPath(proto.FileTypePipelineConfig, path) {
		return fmt.Errorf("The pipeline name(%s) is exists", path)
	}

	err := os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0640)
	if err != nil {
		return err
	}

	_, err = s.pipelineManager.AddPipeline(conf)
	if err != nil {
		return err
	}

	err = s.metadata.AddPath(proto.FileTypePipelineConfig, path)
	if err != nil {
		return err
	}
	return nil
}

func (s *pipelineService) Find(name string) (*proto.PipelineView, error) {
	pipe := s.pipelineManager.Find(name)
	if pipe == nil {
		return nil, errors.New("Not found pipeline " + name)
	}
	return convertPipeliner2PipelineView(pipe), nil
}

func (s *pipelineService) Recreate(conf pipeline.Config) error {
	pipe, err := s.pipelineManager.RecreatePipeline(conf)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(pipe.GetConfig())
	if err != nil {
		return err
	}

	path := s.getConfigPath(pipe.Name())

	return s.metadata.Overwrite(proto.FileTypePipelineConfig, path, data)
}

func (s *pipelineService) List() ([]proto.PipelineView, error) {
	var res []proto.PipelineView
	for _, p := range s.pipelineManager.List() {
		res = append(res, *convertPipeliner2PipelineView(p))
	}
	return res, nil
}

func (s *pipelineService) Control(cmd proto.ControlCommand, names []string) error {
	methodMap := map[proto.ControlCommand]func(names ...string) error{
		proto.ControlCommandStart:   s.pipelineManager.Start,
		proto.ControlCommandStop:    s.pipelineManager.Stop,
		proto.ControlCommandRestart: s.pipelineManager.Restart,
	}

	m, ok := methodMap[cmd]
	if !ok {
		return errors.New("Unsupported method " + string(cmd))
	}

	err := m(names...)
	if err != nil {
		return err
	}
	return nil
}

func (s *pipelineService) getConfigPath(name string) string {
	if !strings.HasSuffix(name, defaultConfigSuffix) {
		name = name + defaultConfigSuffix
	}
	path := s.metadata.GetPath(proto.FileTypePipelineConfig, name)
	return path
}

func convertPipeliner2PipelineView(p pipeline.Pipeliner) *proto.PipelineView {
	return &proto.PipelineView{
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
		Components:    convertComponents(p.ListComponents()),
		Processors:    convertProcessors(p.ListProcessors()),
		RawConfig:     mustMarshalConfig(p.GetConfig()),
	}
}

func mustMarshalConfig(config pipeline.Config) []byte {
	b, _ := yaml.Marshal(config)
	return b
}

func convertComponents(comps []pipeline.Component) []proto.ComponentView {
	var res []proto.ComponentView
	for _, c := range comps {
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

	return res
}

func convertProcessors(procs []pipeline.Processor) []proto.ProcessorView {
	var res []proto.ProcessorView
	for _, c := range procs {
		res = append(res, proto.ProcessorView{
			Name:         c.Name,
			RawConfig:    c.RawConfig,
			Description:  c.Factory.Description(),
			SampleConfig: c.Factory.SampleConfig(),
		})
	}
	return res
}
