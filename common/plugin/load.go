package plugin

import (
	"errors"
	goplugin "plugin"
	"sync"
	"time"

	"github.com/shima-park/nezha/common/log"
)

var (
	rwlock        sync.RWMutex
	loadedPlugins []Plugin
)

type Plugin struct {
	Path     string
	Module   string
	OpenTime time.Time
}

func List() []Plugin {
	rwlock.RLock()
	defer rwlock.RUnlock()

	var snapshots []Plugin
	for _, p := range loadedPlugins {
		snapshots = append(snapshots, p)
	}
	return snapshots
}

func LoadPlugins(path string) error {
	return loadPlugins(path)
}

func loadPlugins(path string) error {
	rwlock.Lock()
	defer rwlock.Unlock()

	log.Info("loading plugin bundle: %v", path)

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

		loadedPlugins = append(loadedPlugins, Plugin{
			Path:     path,
			Module:   name,
			OpenTime: time.Now(),
		})
	}

	return nil
}
