package redis

import (
	"reflect"

	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/common/log"
	"github.com/shima-park/nezha/component"

	"github.com/go-redis/redis"
)

var (
	factory       component.Factory   = NewFactory()
	_             component.Component = &Client{}
	defaultConfig                     = Config{
		Name:     "MyRedisClient",
		Addr:     "127.0.0.1:18000",
		Password: "",
		DB:       0,
		PoolSize: 5,
	}
	description = "redis client factory"
)

func init() {
	if err := component.Register("redis_client", factory); err != nil {
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
	var conf Config
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Redis config: %+v", conf)

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

func (c *Client) Instance() component.Instance {
	return c.instance
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return c.c.Close()
}
