package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/shima-park/lotus/common/log"
	"gopkg.in/yaml.v2"
)

const (
	METADATA_PATH     = "meta"
	METADATA_FILENAME = "meta.yaml"
)

type FileType string

const (
	FileTypePlugin         FileType = "plugins"
	FileTypePipelineConfig FileType = "pipelines"
)

type Metadata interface {
	AddPath(ft FileType, path string) error
	RemovePath(ft FileType, path string) error
	GetPath(ft FileType, filename string) string
	ExistsPath(ft FileType, path string) bool
	ListPaths(ft FileType) []string
}

type metadata struct {
	metapath string
	metafile string

	lock  sync.RWMutex
	paths map[FileType][]string
}

func NewMetadata(metapath string) (Metadata, error) {
	if metapath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		metapath = filepath.Join(pwd, METADATA_PATH)
	}

	m := &metadata{
		metapath: metapath,
		metafile: filepath.Join(metapath, METADATA_FILENAME),
		paths:    map[FileType][]string{},
	}

	err := os.MkdirAll(metapath, 0750)
	if err != nil {
		return nil, fmt.Errorf("Failed to create data path %s: %v", metapath, err)
	}

	_, err = os.Lstat(m.metafile)
	if err != nil && os.IsNotExist(err) {
		log.Info("No metadata file found under: %s. Creating a new registry file.", m.metafile)
		if err = m.save(); err != nil {
			return nil, err
		}
	} else {
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadFile(m.metafile)
		if err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(data, m.paths); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *metadata) save() error {
	data, err := yaml.Marshal(m.paths)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.metafile, data, 0644)
}

func (m *metadata) AddPath(ft FileType, path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	log.Info("Add file type: %s path: %s", ft, path)

	_, ok := m.paths[ft]
	if !ok {
		m.paths[ft] = []string{}
	}

	switch ft {
	case FileTypePlugin:
		return m.addPath(path, "*.so", ft)
	case FileTypePipelineConfig:
		return m.addPath(path, "*.yaml", ft)
	default:
		return fmt.Errorf("Unknown file type: %s", ft)
	}
}

func (m *metadata) GetPath(ft FileType, filename string) string {
	switch ft {
	case FileTypePlugin:
		return filepath.Join(m.metapath, string(FileTypePlugin), filename)
	case FileTypePipelineConfig:
		return filepath.Join(m.metapath, string(FileTypePipelineConfig), filename)
	default:
		panic(fmt.Sprintf("Unknown file type: %s", ft))
	}
}

func (m *metadata) RemovePath(ft FileType, path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, s := range m.paths[ft] {
		if s == path {
			m.paths[ft] = append(m.paths[ft][:i], m.paths[ft][i+1:]...)
			break
		}
	}

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return m.save()
}

func (m *metadata) ExistsPath(ft FileType, path string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, s := range m.paths[ft] {
		if s == path {
			return true
		}
	}
	return false
}

func (m *metadata) ListPaths(ft FileType) []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	paths, ok := m.paths[ft]
	if ok {
		return paths
	}

	return nil
}

func (m *metadata) addPath(path string, pattern string, ft FileType) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	var paths []string
	if fi.IsDir() {
		var err error
		paths, err = filepath.Glob(filepath.Join(path, pattern))
		if err != nil {
			return err
		}

		if len(paths) == 0 {
			return errors.New("not match any " + pattern)
		}
	} else {
		paths = []string{path}
	}

	for _, p := range paths {
		if stringInSlice(p, m.paths[ft]) {
			continue
		}

		m.paths[ft] = append(m.paths[ft], p)
	}

	return m.save()
}

func stringInSlice(t string, ss []string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}
