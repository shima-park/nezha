package client

import "github.com/shima-park/nezha/pkg/rpc/proto"

type component struct {
	apiBuilder
}

func (c *component) List() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	err := GetJSON(c.api("/component/list"), &res)
	return res, err
}

func (c *component) Config(name string) (string, error) {
	var s string
	err := GetJSON(c.api("/component/config?name="+name), &s)
	return s, err
}
