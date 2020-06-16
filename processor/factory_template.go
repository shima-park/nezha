package processor

import (
	"fmt"
	"reflect"

	"github.com/shima-park/nezha/common/config"
)

type FactoryTemplate struct {
	sampleConfig string
	description  string
	factoryFunc  FactoryFunc
}

func NewFactory(sampleConfig interface{}, description string, factoryFunc FactoryFunc) Factory {
	var conf string
	if sampleConfig != nil {
		t := reflect.TypeOf(sampleConfig)
		if t.Kind() == reflect.String {
			conf = fmt.Sprint(sampleConfig)
		} else {
			conf = config.MustMarshal(sampleConfig)
		}
	}

	return FactoryTemplate{
		sampleConfig: conf,
		description:  description,
		factoryFunc:  factoryFunc,
	}
}

func NewFactoryWithProcessor(sampleConfig interface{}, description string, p Processor) Factory {
	return NewFactory(sampleConfig, description, func(string) (Processor, error) {
		return p, nil
	})
}

func (f FactoryTemplate) SampleConfig() string {
	return f.sampleConfig
}

func (f FactoryTemplate) Description() string {
	return f.description
}

func (f FactoryTemplate) New(config string) (Processor, error) {
	return f.factoryFunc(config)
}
