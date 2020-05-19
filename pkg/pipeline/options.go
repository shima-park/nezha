package pipeline

import (
	"context"

	"github.com/google/uuid"
	"github.com/shima-park/nezha/pkg/component"

	"github.com/shima-park/inject"
)

type Option func(*Pipeline)

func WithContext(ctx context.Context) Option {
	return func(p *Pipeline) {
		p.ctx = ctx
	}
}

func WithComponents(components ...component.Component) Option {
	return func(p *Pipeline) {
		p.components = components
	}
}

func WithInjector(injector inject.Injector) Option {
	return func(p *Pipeline) {
		p.injector = injector
	}
}

func WithStream(stream *Stream) Option {
	return func(p *Pipeline) {
		p.stream = stream
	}
}

func WithName(name string) Option {
	return func(p *Pipeline) {
		p.name = name
	}
}

func apply(p *Pipeline, opts []Option) {
	for _, opt := range opts {
		opt(p)
	}

	if p.name == "" {
		p.name = uuid.New().String()
	}
}
