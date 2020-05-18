package component

import (
	"errors"
	"nezha/pkg/common/plugin"
)

type componentPlugin struct {
	name    string
	factory Factory
}

const pluginKey = "component"

func init() {
	plugin.MustRegisterLoader(pluginKey, func(ifc interface{}) error {
		p, ok := ifc.(componentPlugin)
		if !ok {
			return errors.New("plugin does not match handler plugin type")
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
	return plugin.MakePlugin(pluginKey, componentPlugin{module, factory})
}
