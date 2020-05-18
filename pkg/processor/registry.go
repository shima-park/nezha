package processor

import (
	"fmt"
	"github.com/shima-park/nezha/pkg/common/log"
)

type Factory = func(config string) (Processor, error)

var registry = make(map[string]Factory)

func Register(name string, factory Factory) error {
	log.Info("Registering processor factory: %s", name)
	if name == "" {
		return fmt.Errorf("Error registering processor: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering processor '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering processor '%v': already registered", name)
	}

	registry[name] = factory
	log.Info("Successfully registered processor: %s", name)

	return nil
}

func GetFactory(name string) (Factory, error) {
	if _, exists := registry[name]; !exists {
		return nil, fmt.Errorf("Error creating processor. No such processor type exist: '%v'", name)
	}
	return registry[name], nil
}

func New(name string, config string) (Processor, error) {
	f, err := GetFactory(name)
	if err != nil {
		return nil, err
	}

	h, err := f(config)
	if err != nil {
		return nil, err
	}

	err = Validate(h)
	if err != nil {
		return nil, err
	}
	return h, nil
}
