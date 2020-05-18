package component

import (
	"fmt"
	"github.com/shima-park/nezha/pkg/common/log"
)

type Factory = func(config string) (Component, error)

var registry = make(map[string]Factory)

func Register(name string, factory Factory) error {
	log.Info("Registering component factory: %s", name)
	if name == "" {
		return fmt.Errorf("Error registering component: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering component '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering component '%v': already registered", name)
	}

	registry[name] = factory
	log.Info("Successfully registered component: %s", name)

	return nil
}

func GetFactory(name string) (Factory, error) {
	if _, exists := registry[name]; !exists {
		return nil, fmt.Errorf("Error creating component. No such component type exist: '%v'", name)
	}
	return registry[name], nil
}
