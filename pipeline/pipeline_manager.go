package pipeline

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/common/log"
)

type PipelinerManager interface {
	AddPipeline(config Config) (Pipeliner, error)
	RemovePipeline(name ...string) error
	List() []Pipeliner
	Find(name string) Pipeliner
	Restart(name ...string) error
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

		_, err = pm.AddPipeline(conf)
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
	defer p.rwlock.Unlock()

	return p.find(name)
}

func (p *pipelinerManager) find(name string) Pipeliner {
	return p.pipelines[name]
}

func (p *pipelinerManager) AddPipeline(config Config) (Pipeliner, error) {
	p.rwlock.Lock()
	defer p.rwlock.Unlock()

	return p.addPipeline(config)
}

func (p *pipelinerManager) addPipeline(config Config) (Pipeliner, error) {
	_, ok := p.pipelines[config.Name]
	if ok {
		return nil, fmt.Errorf("Pipeline: %s is already register", config.Name)
	}

	pipe, err := NewPipelineByConfig(config)
	if err != nil {
		return nil, err
	}
	p.pipelines[config.Name] = pipe
	return pipe, nil
}

func (p *pipelinerManager) RemovePipeline(names ...string) error {
	return p.doByName(false, names, p.removePipeline)
}

func (p *pipelinerManager) removePipeline(pipe Pipeliner) error {
	pipe.Stop()
	delete(p.pipelines, pipe.Name())
	return nil
}

func (p *pipelinerManager) Restart(names ...string) error {
	return p.doByName(false, names, func(oldPipe Pipeliner) error {
		name := oldPipe.Name()
		err := p.removePipeline(oldPipe)
		if err != nil {
			return errors.Wrap(err, name)
		}

		pipe, err := p.addPipeline(oldPipe.GetConfig())
		if err != nil {
			return errors.Wrap(err, name)
		}

		err = pipe.Start()
		if err != nil {
			return errors.Wrap(err, name)
		}
		return nil
	})
}

func (p *pipelinerManager) Start(names ...string) error {
	return p.doByName(true, names, func(pipe Pipeliner) error {
		if pipe.State() == Exited {
			return fmt.Errorf("Pipeline(%s)'s state is exited, please try to restart it", pipe.Name())
		}
		return pipe.Start()
	})
}

func (p *pipelinerManager) Stop(names ...string) error {
	return p.doByName(true, names, func(pipe Pipeliner) error {
		pipe.Stop()
		return nil
	})
}

func (p *pipelinerManager) doByName(isReadLock bool, names []string, callback func(pipe Pipeliner) error) error {
	if isReadLock {
		p.rwlock.RLock()
		defer p.rwlock.RUnlock()
	} else {
		p.rwlock.Lock()
		defer p.rwlock.Unlock()
	}

	var errs []string
	for _, name := range names {
		pipe := p.find(name)
		if pipe == nil {
			errs = append(errs, fmt.Sprintf("Pipeline: %s is not found", name))
			continue
		}

		err := callback(pipe)
		if err != nil {
			errs = append(errs, errors.Wrap(err, name).Error())
			continue
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ""))
	}

	return nil
}
