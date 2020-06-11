package redis

import (
	"reflect"

	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"

	"github.com/go-redis/redis"
)

var _ component.Component = &Client{}

func init() {
	if err := component.Register("redis_client", func(config string) (component.Component, error) {
		return NewClient(config)
	}); err != nil {
		panic(err)
	}
}

type ClientConfig struct {
	Name     string `yaml:"name"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
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

	log.Info("Pika config: %+v", conf)

	c := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

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

func (c *Client) SampleConfig() string {
	conf := ClientConfig{
		Name:     "MyRedisClient",
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
