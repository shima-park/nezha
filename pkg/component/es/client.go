package es

import (
	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/component"
	"reflect"

	"github.com/olivere/elastic"
)

var _ component.Component = &Client{}

func init() {
	if err := component.Register("es_client", func(config string) (component.Component, error) {
		return NewWriter(config)
	}); err != nil {
		panic(err)
	}
}

type ClientConfig struct {
	Name string
	Addr string
}

type Client struct {
	c        *elastic.Client
	instance component.Instance
}

func NewClient(rawConfig string) (*Client, error) {
	var conf ClientConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	var options elastic.ClientOptionFunc
	if conf.Addr != "" {
		options = append(options, elastic.SetURL(conf.Addr))
	}

	return &Client{
		c: elastic.NewClient(options...),
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(c.c),
			reflect.ValueOf(c.c),
			c.c,
		),
	}, nil
}

func (c *Client) SampleConfig() string {
	conf := ClientConfig{
		Name: "MyES",
		Addr: "127.0.0.1:9200",
	}

	b, _ := config.Marshal(&conf)

	return string(b)
}

func (c *Client) Description() string {
	return "es client factory"
}

func (c *Client) Instance() Instance {
	return c.instance
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return nil
}