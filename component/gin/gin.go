package gin

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/common/config"
	"github.com/shima-park/nezha/common/log"
	"github.com/shima-park/nezha/component"
)

var (
	factory                    component.Factory   = NewFactory()
	_                          component.Component = &Gin{}
	defaultGracefulStopTimeout                     = time.Second * 30
	defaultConfig                                  = Config{
		Name:                "GinServer",
		Addr:                ":8080",
		GracefulStopTimeout: defaultGracefulStopTimeout,
	}
	description = "http listen factory"
)

func init() {
	if err := component.Register("gin_server", factory); err != nil {
		panic(err)
	}
}

func NewFactory() component.Factory {
	return component.NewFactory(
		defaultConfig,
		description,
		func(c string) (component.Component, error) {
			return NewGin(c)
		})
}

type Config struct {
	Name                string        `yaml:"name"`
	Addr                string        `yaml:"addr"`
	GracefulStopTimeout time.Duration `yaml:"graceful_stop_timeout"`
}

type Gin struct {
	conf     Config
	srv      *http.Server
	gin      *gin.Engine
	instance component.Instance
}

func NewGin(rawConfig string) (*Gin, error) {
	conf := defaultConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Gin config: %+v", conf)

	if conf.Name == "" {
		return nil, errors.New("Component:gin_server name cannot be empty")
	}

	if conf.Addr == "" {
		return nil, errors.New("Component:gin_server name cannot be empty")
	}

	if conf.GracefulStopTimeout < time.Second {
		conf.GracefulStopTimeout = defaultGracefulStopTimeout
	}

	g := gin.New()

	srv := &http.Server{
		Addr:    conf.Addr,
		Handler: g,
	}

	return &Gin{
		srv: srv,
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(g),
			reflect.ValueOf(g),
			g,
		),
	}, nil
}

func (g *Gin) Instance() component.Instance {
	return g.instance
}

func (g *Gin) Start() error {
	var errCh = make(chan error)
	go func() {
		// service connections
		if err := g.srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(time.Second):
		return nil
	}

	return nil
}

func (g *Gin) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), g.conf.GracefulStopTimeout)
	defer cancel()
	if err := g.srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
