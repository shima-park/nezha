package pika

import (
	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/component"
	"reflect"

	"github.com/go-redis/redis"
)

var _ component.Component = &Client{}

func init() {
	if err := component.Register("pika_client", func(config string) (component.Component, error) {
		return NewWriter(config)
	}); err != nil {
		panic(err)
	}
}

type ClientConfig struct {
	Name     string
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type Client struct {
	c        *redis.Client
	instance component.Instance
}

func NewClient(rawConfig string) (*Client, error) {
	var conf ClientConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: redis.NewClient(&redis.Options{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.DB,
			PoolSize: conf.PoolSize,
		}),
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
		Name:     "MyPiKa",
		Addr:     "127.0.0.1:18000",
		Password: "",
		DB:       0,
		PoolSize: 5,
	}

	b, _ := config.Marshal(&conf)

	return string(b)
}

func (c *Client) Description() string {
	return "pika client factory"
}

func (c *Client) Instance() component.Instance {
	return c.instance
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return c.c.Close()
}
