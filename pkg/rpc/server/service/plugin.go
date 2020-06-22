package service

import (
	"fmt"

	"github.com/shima-park/lotus/common/plugin"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

type pluginService struct {
	metadata proto.Metadata
}

func NewPluginService(metadata proto.Metadata) proto.Plugin {
	return &pluginService{
		metadata: metadata,
	}
}

func (s *pluginService) List() ([]proto.PluginView, error) {
	var res []proto.PluginView
	for _, p := range plugin.List() {
		res = append(res, proto.PluginView{
			Path:     p.Path,
			Module:   p.Module,
			OpenTime: fmt.Sprint(p.OpenTime),
		})
	}
	return res, nil
}

func (s *pluginService) Open(path string) error {
	return plugin.LoadPlugins(path)
}

func (s *pluginService) Add(path string) error {
	err := s.metadata.AddPath(proto.FileTypePlugin, path)
	if err != nil {
		return err
	}

	err = plugin.LoadPlugins(path)
	if err != nil {
		_ = s.metadata.RemovePath(proto.FileTypePlugin, path)
		return err
	}
	return nil
}
