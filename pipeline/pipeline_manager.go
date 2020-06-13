package pipeline

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/common/log"
)

type PipelinerManager interface {
	AddPipeline(config Config) error
	List() []Pipeliner
	Find(name string) Pipeliner
	Start(name ...string) error
	Stop(name ...string) error
}

type pipelinerManager struct {
	rwlock    sync.RWMutex
	pipelines map[string]Pipeliner // key: name value: Pipeliner
}

func NewPipelinerManager() PipelinerManager {
	pm := &pipelinerManager{
		pipelines: map[string]Pipeliner{},
	}
	return pm
}

func NewPipelineManagerByConfigs(path string) (PipelinerManager, error) {
	pm := NewPipelinerManager()
	return pm, loadPipelineFromFile(path, pm)
}

func loadPipelineFromFile(path string, pm PipelinerManager) error {
	log.Info("Read config from: %s", path)

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	var paths []string
	if fi.IsDir() {
		var err error
		paths, err = filepath.Glob(filepath.Join(path, "*.yaml"))
		if err != nil {
			return err
		}

		if len(paths) == 0 {
			return errors.New("not match pipeline config")
		}
	} else {
		paths = []string{path}
	}

	for _, path := range paths {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var conf Config
		err = config.Unmarshal(content, &conf)
		if err != nil {
			return err
		}

		err = pm.AddPipeline(conf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *pipelinerManager) List() []Pipeliner {
	var ps []Pipeliner

	p.rwlock.RLock()
	for _, p := range p.pipelines {
		ps = append(ps, p)
	}
	p.rwlock.RUnlock()

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name() < ps[j].Name()
	})

	return ps
}

func (p *pipelinerManager) Find(name string) Pipeliner {
	p.rwlock.Lock()
	pipe, ok := p.pipelines[name]
	if ok {
		p.rwlock.Unlock()
		return pipe
	}
	p.rwlock.Unlock()

	return nil
}

func (p *pipelinerManager) AddPipeline(config Config) error {
	pipe, err := NewPipelineByConfig(config)
	if err != nil {
		return err
	}

	p.rwlock.Lock()
	p.pipelines[pipe.Name()] = pipe
	p.rwlock.Unlock()
	return nil
}

func (p *pipelinerManager) Start(names ...string) error {
	var startFuncs []func() error
	p.rwlock.RLock()
	for _, name := range names {
		pipeline := p.Find(name)
		if pipeline == nil {
			return errors.New("Not found pipeline: " + name)
		}
		startFuncs = append(startFuncs, pipeline.Start)
	}
	p.rwlock.RUnlock()

	for _, start := range startFuncs {
		err := start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *pipelinerManager) Stop(names ...string) error {
	var stopFuncs []func()
	p.rwlock.RLock()
	for _, name := range names {
		pipeline := p.Find(name)
		if pipeline == nil {
			return errors.New("Not found pipeline: " + name)
		}
		stopFuncs = append(stopFuncs, pipeline.Stop)
	}
	p.rwlock.RUnlock()

	for _, stop := range stopFuncs {
		stop()
	}
	return nil
}
