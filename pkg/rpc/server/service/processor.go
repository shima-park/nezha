package service

import (
	"sort"

	"github.com/shima-park/lotus/processor"
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

type processorService struct {
}

func NewProcessorService() proto.Processor {
	return &processorService{}
}

func (s *processorService) List() ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	for _, c := range processor.ListFactory() {
		res = append(res, *newProcessorView(c.Name, c.Factory))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (s *processorService) Find(name string) (*proto.ProcessorView, error) {
	factory, err := processor.GetFactory(name)
	if err != nil {
		return nil, err
	}

	return newProcessorView(name, factory), nil
}

func newProcessorView(name string, factory processor.Factory) *proto.ProcessorView {
	return &proto.ProcessorView{
		Name:         name,
		SampleConfig: factory.SampleConfig(),
		Description:  factory.Description(),
	}
}
