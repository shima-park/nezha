package service

import (
	"sort"

	"github.com/shima-park/lotus/component"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

type componentService struct {
}

func NewComponentService() proto.Component {
	return &componentService{}
}

func (s *componentService) List() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	for _, c := range component.ListFactory() {
		res = append(res, *newComponentView(c.Name, c.Factory))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (s *componentService) Find(name string) (*proto.ComponentView, error) {
	factory, err := component.GetFactory(name)
	if err != nil {
		return nil, err
	}

	return newComponentView(name, factory), nil
}

func newComponentView(name string, factory component.Factory) *proto.ComponentView {
	return &proto.ComponentView{
		Name:         name,
		SampleConfig: factory.SampleConfig(),
		Description:  factory.Description(),
	}
}
