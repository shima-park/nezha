package client

import "github.com/shima-park/nezha/pkg/rpc/proto"

type processor struct {
	apiBuilder
}

func (c *processor) List() ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	err := GetJSON(c.api("/processor/list"), &res)
	return res, err
}

func (c *processor) Find(name string) (*proto.ProcessorView, error) {
	var res proto.ProcessorView
	err := GetJSON(c.api("/processor?name="+name), &res)
	return &res, err
}
