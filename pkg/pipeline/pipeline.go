package pipeline

import (
	"context"
	"errors"
	"fmt"
	"nezha/pkg/common/log"
	"nezha/pkg/component"

	"github.com/google/uuid"
	"github.com/shima-park/inject"
)

type Pipeline struct {
	ctx        context.Context
	id         string
	name       string
	components []component.Component
	injector   inject.Injector
	stream     *Stream
	done       chan struct{}
}

func New(opts ...Option) (*Pipeline, error) {
	p := &Pipeline{
		ctx:      context.Background(),
		id:       uuid.New().String(),
		injector: inject.New(),
	}

	apply(p, opts)

	if p.name == "" {
		p.name = p.id
	}

	if p.stream == nil {
		return nil, errors.New("The pipeline must have at least one stream")
	}

	for _, component := range p.components {
		instance := component.Instance()
		p.injector.Set(instance.Type, instance.Name, instance.Value)
	}

	return p, nil
}

func NewPipelineByConfig(conf Config, opts ...Option) (*Pipeline, error) {
	var components []component.Component
	for cn, rawConfig := range conf.Components {
		factory, err := component.GetFactory(cn)
		if err != nil {
			return nil, err
		}
		c, err := factory(rawConfig)
		if err != nil {
			return nil, err
		}
		components = append(components, c)
	}

	stream, err := NewStream(*conf.Stream)
	if err != nil {
		return nil, err
	}

	return New(
		append(
			[]Option{
				WithComponents(components...),
				WithStream(stream),
			},
			opts...,
		)...,
	)
}

func (c *Pipeline) ID() string {
	return c.id
}

func (c *Pipeline) Name() string {
	return c.name
}

func (c *Pipeline) Start() error {
	for !c.isStopped() {
		select {
		case <-c.ctx.Done():
			c.Stop()
		default:
		}

		ctx, cancel := context.WithCancel(c.ctx)
		defer cancel()

		c := NewExecContext(ctx, c.stream, c.injector)
		if err := c.Run(); err != nil {
			log.Error("[Pipeline] error: %v", err)
		}
	}
	return nil
}

func (c *Pipeline) isStopped() bool {
	select {
	case <-c.done:
		return true
	default:
	}
	return false
}

func (c *Pipeline) Stop() {
	if c.isStopped() {
		return
	}
	close(c.done)
}

func (c *Pipeline) ListComponent() []string {
	var list []string
	for _, component := range c.components {
		i := component.Instance()
		list = append(list, fmt.Sprintf("Name: %s, Type: %s", i.Name, i.Type))
	}
	return list
}
