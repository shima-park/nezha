package gin

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"
)

var _ component.Component = &Gin{}

func init() {
	if err := component.Register("gin_server", func(config string) (component.Component, error) {
		return NewGin(config)
	}); err != nil {
		panic(err)
	}
}

type GinConfig struct {
	Name    string            `yaml:"name"`
	Addr    string            `yaml:"addr"`
	Routers map[string]string `yaml:"routers"`
}

type Gin struct {
	conf     GinConfig
	srv      *http.Server
	ctxCh    chan *gin.Context
	instance component.Instance
}

func NewGin(rawConfig string) (*Gin, error) {
	var conf GinConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Gin config: %+v", conf)

	g := gin.New()

	var ctxCh = make(chan *gin.Context, 1024)
	handler := func(ctx *gin.Context) {
		select {
		case ctxCh <- ctx:
		default:
			ctx.Abort()
		}
	}

	for method, path := range conf.Routers {
		switch strings.ToUpper(method) {
		case "Any":
			g.Any(path, handler)
		case "GET":
			g.GET(path, handler)
		case "POST":
			g.POST(path, handler)
		case "DELETE":
			g.DELETE(path, handler)
		case "PATCH":
			g.PATCH(path, handler)
		case "PUT":
			g.PUT(path, handler)
		case "OPTIONS":
			g.OPTIONS(path, handler)
		case "HEAD":
			g.HEAD(path, handler)
		default:
			return nil, fmt.Errorf("Unsupported method: %s path: %s", method, path)
		}
	}

	srv := &http.Server{
		Addr:    conf.Addr,
		Handler: g,
	}

	return &Gin{
		srv:   srv,
		ctxCh: ctxCh,
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(ctxCh),
			reflect.ValueOf(ctxCh),
			ctxCh,
		),
	}, nil
}

func (g *Gin) SampleConfig() string {
	conf := GinConfig{
		Name: "GinServer",
		Addr: ":8080",
		Routers: map[string]string{
			"GET": "/hello",
		},
	}

	b, _ := config.Marshal(&conf)

	return string(b)
}

func (g *Gin) Description() string {
	return "http listen factory"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := g.srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
