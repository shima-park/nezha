package plugin

import (
	"errors"
	"flag"
	goplugin "plugin"
	"strings"

	"github.com/shima-park/nezha/common/log"
)

var plugins = &pluginList{}

func init() {
	flag.Var(plugins, "plugin", "Load additional plugins")
}

type pluginList struct {
	paths []string
}

func (p *pluginList) String() string {
	return strings.Join(p.paths, ",")
}

func (p *pluginList) Set(v string) error {
	for _, path := range p.paths {
		if path == v {
			log.Warn("%s is already a registered plugin", path)
			return nil
		}
	}
	p.paths = append(p.paths, v)
	return nil
}

func Initialize() error {
	for _, path := range plugins.paths {
		log.Info("loading plugin bundle: %v", path)

		if err := loadPlugins(path); err != nil {
			return err
		}
	}

	return nil
}

func loadPlugins(path string) error {
	p, err := goplugin.Open(path)
	if err != nil {
		return err
	}

	sym, err := p.Lookup("Bundle")
	if err != nil {
		return err
	}

	ptr, ok := sym.(*map[string][]interface{})
	if !ok {
		return errors.New("invalid bundle type")
	}

	bundle := *ptr
	for name, plugins := range bundle {
		loader := registry[name]
		if loader == nil {
			continue
		}

		for _, plugin := range plugins {
			if err := loader(plugin); err != nil {
				return err
			}
		}
	}

	return nil
}
