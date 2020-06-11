package pipeline

import (
	"github.com/shima-park/nezha/pkg/component"
	"github.com/shima-park/nezha/pkg/processor"
)

type Config struct {
	Name       string              `yaml:"name"`
	Components []map[string]string `yaml:"components"` // key: name, value: rawConfig
	Processors []map[string]string `yaml:"processors"` // key: name, value: rawConfig
	Pipeline   PipelineConfig      `yaml:"pipeline"`
}

type PipelineConfig struct {
	Schedule string        `yaml:"schedule"`
	Stream   *StreamConfig `yaml:"stream"` // key: name, value: StreamConfig
}

func (c *Config) NewComponents() ([]NamedComponent, error) {
	var components []NamedComponent
	for _, name2config := range c.Components {
		for componentName, rawConfig := range name2config {
			factory, err := component.GetFactory(componentName)
			if err != nil {
				return nil, err
			}
			c, err := factory(rawConfig)
			if err != nil {
				return nil, err
			}
			components = append(components, NamedComponent{
				Name:      componentName,
				RawConfig: rawConfig,
				Component: c,
			})
		}
	}

	return components, nil
}

func (c *Config) NewProcessors() ([]NamedProcessor, error) {
	var processors []NamedProcessor
	for _, name2config := range c.Processors {
		for processorName, rawConfig := range name2config {
			factory, err := processor.GetFactory(processorName)
			if err != nil {
				return nil, err
			}

			p, err := factory(rawConfig)
			if err != nil {
				return nil, err
			}
			processors = append(processors, NamedProcessor{
				Name:      processorName,
				RawConfig: rawConfig,
				Processor: p,
			})
		}
	}
	return processors, nil
}
