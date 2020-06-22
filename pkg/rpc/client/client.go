package client

import (
	"github.com/shima-park/nezha/pkg/rpc/proto"
)

type Client struct {
	proto.Pipeline
	proto.Component
	proto.Processor
	proto.Plugin
	proto.Server

	addr string
}

func NewClient(addr string) *Client {
	addr = normalizeURL(addr)
	b := apiBuilder{addr}
	return &Client{
		Pipeline:  &pipeline{b},
		Component: &component{b},
		Processor: &processor{b},
		Plugin:    &plugin{b},
		Server:    &server{b},
	}
}

type apiBuilder struct {
	addr string
}

func (b *apiBuilder) api(path string) string {
	return b.addr + path
}
