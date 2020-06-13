package pipeline

import (
	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/component"
	"github.com/shima-park/nezha/processor"
)

type Config struct {
	Name       string              `yaml:"name"`
	Schedule   string              `yaml:"schedule"`   // 调度计划，为空时死循环调度，可以传入cron表达式调度
	Bootstrap  bool                `yaml:"bootstrap"`  // 随进程启动而启动
	Components []map[string]string `yaml:"components"` // key: name, value: rawConfig
	Processors []map[string]string `yaml:"processors"` // key: name, value: rawConfig
	Stream     StreamConfig        `yaml:"stream"`     // key: name, value: StreamConfig
}

func (c Config) NewComponents() ([]Component, error) {
	var components []Component
	for _, name2config := range c.Components {
		for componentName, rawConfig := range name2config {
			factory, err := component.GetFactory(componentName)
			if err != nil {
				return nil, err
			}
			c, err := factory.New(rawConfig)
			if err != nil {
				return nil, err
			}
			components = append(components, Component{
				Name:      componentName,
				RawConfig: rawConfig,
				Component: c,
				Factory:   factory,
			})
		}
	}

	return components, nil
}

func (c Config) NewProcessors() ([]Processor, error) {
	var processors []Processor
	for _, name2config := range c.Processors {
		for processorName, rawConfig := range name2config {
			factory, err := processor.GetFactory(processorName)
			if err != nil {
				return nil, err
			}

			p, err := factory.New(rawConfig)
			if err != nil {
				return nil, err
			}
			processors = append(processors, Processor{
				Name:      processorName,
				RawConfig: rawConfig,
				Processor: p,
			})
		}
	}
	return processors, nil
}

func (c Config) Marshal() ([]byte, error) {
	return config.Marshal(c)
}
