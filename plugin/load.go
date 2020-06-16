package plugin

import (
	"errors"
	"os"
	"path/filepath"
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
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	var paths []string
	if fi.IsDir() {
		var err error
		paths, err = filepath.Glob(filepath.Join(path, "*.so"))
		if err != nil {
			return err
		}

		if len(paths) == 0 {
			return errors.New("not match any *.so for plugin")
		}
	} else {
		paths = []string{path}
	}

	for _, path := range paths {
		err := loadPlugins(path)
		if err != nil {
			return err
		}
	}
	return nil
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
