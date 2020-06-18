package client

import (
	"net/url"

	p "github.com/shima-park/nezha/pipeline"
	"github.com/shima-park/nezha/rpc/proto"
)

type pipeline struct {
	apiBuilder
}

func (p *pipeline) List() ([]proto.PipelineView, error) {
	var res []proto.PipelineView
	err := GetJSON(p.api("/pipeline/list"), &res)
	return res, err
}

func (p *pipeline) Add(conf p.Config) error {
	return PostYaml(p.api("/pipeline/add"), conf, nil)
}

func (p *pipeline) Control(cmd proto.ControlCommand, names []string) error {
	vals := url.Values{}
	vals.Add("cmd", string(cmd))
	for _, name := range names {
		vals.Add("name", name)
	}
	return GetJSON(p.api("/pipeline/ctrl"), nil)
}

func (p *pipeline) ListComponents(name string) ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	err := GetJSON(p.api("/pipeline/components?name="+name), &res)
	return res, err
}

func (p *pipeline) ListProcessors(name string) ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	err := GetJSON(p.api("/pipeline/processors?name="+name), &res)
	return res, err
}

func (p *pipeline) Visualize(name string, format proto.VisualizeFormat) error {
	return nil
}

func (p *pipeline) Vars(name string) (map[string]string, error) {
	res := map[string]string{}
	err := GetJSON(p.api("/pipeline/vars?name="+name), &res)
	return res, err
}

func (p *pipeline) Config(name string) (string, error) {
	var s string
	err := GetJSON(p.api("/pipeline/config?name="+name), &s)
	return s, err
}

func (p *pipeline) GenerateConfig(name string, components, processors []string) (string, error) {
	vals := url.Values{}
	vals.Add("name", name)
	for _, c := range components {
		vals.Add("components", c)
	}
	for _, p := range processors {
		vals.Add("processors", p)
	}
	var s string
	err := GetJSON(p.api("/pipeline/generate-config"), &s)
	return s, err
}
