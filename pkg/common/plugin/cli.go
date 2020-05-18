package plugin

import (
	"flag"
	"nezha/pkg/common/log"

	"strings"
)

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

var plugins = &pluginList{}

func init() {
	flag.Var(plugins, "plugin", "Load additional plugins")
}

func Initialize() error {
	if len(plugins.paths) > 0 {
		log.Warn("loadable plugin support is experimental")
	}

	for _, path := range plugins.paths {
		log.Info("loading plugin bundle: %v", path)

		if err := LoadPlugins(path); err != nil {
			return err
		}
	}

	return nil
}
