package client

import "github.com/shima-park/nezha/rpc/proto"

type processor struct {
	apiBuilder
}

func (c *processor) List() ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	err := GetJSON(c.api("/processor/list"), &res)
	return res, err
}

func (c *processor) Config(name string) (string, error) {
	var s string
	err := GetJSON(c.api("/processor/config?name="+name), &s)
	return s, err
}
