package processor

import (
	"errors"

	"github.com/shima-park/nezha/plugin"
)

type processorPlugin struct {
	name    string
	factory Factory
}

const pluginKey = "processor"

func init() {
	plugin.MustRegisterLoader(pluginKey, func(ifc interface{}) error {
		p, ok := ifc.(processorPlugin)
		if !ok {
			return errors.New("plugin does not match processor plugin type")
		}

		if p.factory != nil {
			if err := Register(p.name, p.factory); err != nil {
				return err
			}
		}

		return nil
	})
}

func Plugin(
	module string,
	factory Factory,
) map[string][]interface{} {
	return plugin.MakePlugin(pluginKey, processorPlugin{module, factory})
}
