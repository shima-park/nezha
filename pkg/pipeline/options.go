package pipeline

import (
	"context"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"github.com/shima-park/nezha/pkg/inject"
)

type Option func(*Pipeline)

func WithContext(ctx context.Context) Option {
	return func(p *Pipeline) {
		ctx, cancel := context.WithCancel(ctx)
		p.ctx = ctx
		p.cancel = cancel
	}
}

func WithComponents(components ...NamedComponent) Option {
	return func(p *Pipeline) {
		p.components = components
	}
}

func WithProcessors(processors ...NamedProcessor) Option {
	return func(p *Pipeline) {
		p.processors = processors
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

func WithSchedule(s cron.Schedule) Option {
	return func(p *Pipeline) {
		p.schedule = s
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
