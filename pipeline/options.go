package pipeline

import (
	"context"

	"github.com/robfig/cron/v3"

	"github.com/shima-park/nezha/common/inject"
)

type Option func(*pipeliner)

func WithContext(ctx context.Context) Option {
	return func(p *pipeliner) {
		ctx, cancel := context.WithCancel(ctx)
		p.ctx = ctx
		p.cancel = cancel
	}
}

func WithComponents(components ...Component) Option {
	return func(p *pipeliner) {
		p.components = components
	}
}

func WithProcessors(processors ...Processor) Option {
	return func(p *pipeliner) {
		p.processors = processors
	}
}

func WithInjector(injector inject.Injector) Option {
	return func(p *pipeliner) {
		p.injector = injector
	}
}

func WithStream(stream *Stream) Option {
	return func(p *pipeliner) {
		p.stream = stream
	}
}

func WithName(name string) Option {
	return func(p *pipeliner) {
		p.name = name
	}
}

func WithSchedule(s cron.Schedule) Option {
	return func(p *pipeliner) {
		p.schedule = s
	}
}

func WithConfig(config Config) Option {
	return func(p *pipeliner) {
		p.config = config
	}
}

func apply(p *pipeliner, opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}
