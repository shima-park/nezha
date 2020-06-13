package es

import (
	"reflect"

	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/common/log"
	"github.com/shima-park/nezha/component"

	"github.com/olivere/elastic/v7"
)

var (
	factory       component.Factory   = NewFactory()
	_             component.Component = &Client{}
	defaultConfig                     = Config{
		Name: "MyES",
		Addr: "127.0.0.1:9200",
	}
	description = "es client factory"
)

func init() {
	if err := component.Register("es_client", factory); err != nil {
		panic(err)
	}
}

func NewFactory() component.Factory {
	return component.NewFactory(
		defaultConfig,
		description,
		func(c string) (component.Component, error) {
			return NewClient(c)
		})
}

type Config struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

type Client struct {
	c        *elastic.Client
	instance component.Instance
}

func NewClient(rawConfig string) (*Client, error) {
	conf := defaultConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("ES config: %+v", conf)

	var options []elastic.ClientOptionFunc
	if conf.Addr != "" {
		options = append(options, elastic.SetURL(conf.Addr))
	}

	c, err := elastic.NewClient(options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: c,
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(c),
			reflect.ValueOf(c),
			c,
		),
	}, nil
}

func (c *Client) Instance() component.Instance {
	return c.instance
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return nil
}
