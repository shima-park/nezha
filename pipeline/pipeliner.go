package pipeline

import (
	"fmt"
	"io"

	"context"
	"expvar"

	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/shima-park/nezha/common/log"
	"github.com/shima-park/nezha/component"
	"github.com/shima-park/nezha/inject"
	"github.com/shima-park/nezha/monitor"
	"github.com/shima-park/nezha/processor"
)

type Pipeliner interface {
	Name() string
	Start() error
	Stop()
	State() State
	ListComponents() []Component
	ListProcessors() []Processor
	Monitor() monitor.Monitor
	GetConfig() Config
	SetConfig(config Config) error
	Visualize(w io.Writer, format string) error
	CheckDependence() []error
}

type Component struct {
	Name      string
	RawConfig string
	Component component.Component
	Factory   component.Factory
}

type Processor struct {
	Name      string
	RawConfig string
	Processor processor.Processor
	Factory   processor.Factory
}

type pipeliner struct {
	config Config

	ctx    context.Context
	cancel context.CancelFunc

	name       string
	components []Component
	processors []Processor
	injector   inject.Injector
	stream     *Stream
	monitor    monitor.Monitor
	startTime  time.Time
	schedule   cron.Schedule

	state     int32
	runningWg sync.WaitGroup
}

func New(opts ...Option) (Pipeliner, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &pipeliner{
		ctx:       ctx,
		cancel:    cancel,
		injector:  inject.New(),
		startTime: time.Now(),
	}

	apply(p, opts)

	if p.stream == nil {
		return nil, errors.Wrap(errors.New("The pipeliner must have at least one stream"), p.Name())
	}

	if p.name == "" {
		return nil, errors.Wrap(errors.New("The pipeliner's name cannot be empty "), p.Name())
	}

	p.monitor = monitor.NewMonitor(p.Name())
	p.monitor.Set(METRICS_KEY_PIPELINE_STATE, expvar.Func(func() interface{} { return p.State() }))

	p.injector.MapTo(p.monitor, "Monitor", (*monitor.Monitor)(nil))
	p.injector.MapTo(p.ctx, "Context", (*context.Context)(nil))

	distinct := map[reflect.Type]map[string]struct{}{}
	for _, c := range p.components {
		instance := c.Component.Instance()

		if _, ok := distinct[instance.Type()]; !ok {
			distinct[instance.Type()] = map[string]struct{}{}
		}

		if _, ok := distinct[instance.Type()][instance.Name()]; ok {
			return nil, errors.Wrap(
				errors.Errorf("Type: %s, Name: %s is already registered",
					instance.Type(), instance.Name()),
				"Pipeline: "+p.Name()+" Component: "+c.Name,
			)
		}

		distinct[instance.Type()] = map[string]struct{}{
			instance.Name(): struct{}{},
		}

		p.injector.Set(instance.Type(), instance.Name(), instance.Value())
	}

	if errs := p.CheckDependence(); len(errs) > 0 {
		return nil, errs[0]
	}

	return p, nil
}

func NewPipelineByConfig(conf Config, opts ...Option) (Pipeliner, error) {
	components, err := conf.NewComponents()
	if err != nil {
		return nil, errors.Wrap(err, conf.Name)
	}

	processors, err := conf.NewProcessors()
	if err != nil {
		return nil, errors.Wrap(err, conf.Name)
	}

	var pm = map[string]Processor{}
	for _, p := range processors {
		pm[p.Name] = p
	}

	stream, err := NewStream(conf.Stream, pm)
	if err != nil {
		return nil, errors.Wrap(err, conf.Name)
	}

	var schedule = defaultSchedule
	if conf.Schedule != "" {
		schedule, err = standardParser.Parse(conf.Schedule)
		if err != nil {
			return nil, errors.Wrap(err, conf.Name)
		}
	}

	return New(
		append(
			[]Option{
				WithName(conf.Name),
				WithComponents(components...),
				WithProcessors(processors...),
				WithStream(stream),
				WithSchedule(schedule),
				WithConfig(conf),
			},
			opts...,
		)...,
	)
}

func (p *pipeliner) Name() string {
	return p.name
}

func (p *pipeliner) CheckDependence() []error {
	checkInj := inject.New()
	checkInj.SetParent(p.injector)
	return check(p.stream, checkInj)
}

func (p *pipeliner) newExecContext() *execContext {
	inj := inject.New()
	inj.SetParent(p.injector)

	ctx, cancel := context.WithCancel(p.ctx)
	inj.MapTo(ctx, "Context", (*context.Context)(nil))

	c := &execContext{
		ctx:      ctx,
		cancel:   cancel,
		injector: inj,
		stream:   p.stream,
		monitor:  p.monitor,
		inputC:   make(chan inject.Injector, p.stream.config.BufferSize),
	}

	return c
}

func (p *pipeliner) Start() error {
	if !atomic.CompareAndSwapInt32(&p.state, int32(Idle), int32(Running)) {
		return nil
	}

	for _, c := range p.components {
		if err := c.Component.Start(); err != nil {
			return err
		}
	}

	c := p.newExecContext()
	if err := c.Start(); err != nil {
		return err
	}

	p.runningWg.Add(1)
	go func() {
		defer p.runningWg.Done()

		for !p.isStopped() {
			<-time.After(time.Second)
			p.monitor.Set(METRICS_KEY_PIPELINE_UPTIME, monitor.Elapsed(time.Since(p.startTime)))
		}
	}()

	p.runningWg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Pipeline: %s, Panic: %s, Stack: %s",
					p.Name(), r, string(debug.Stack()))
			}
			c.Stop()

			p.monitor.Set(METRICS_KEY_PIPELINE_EXIT_TIME, monitor.Time(time.Now()))

			p.runningWg.Done()
		}()

		p.monitor.Set(METRICS_KEY_PIPELINE_START_TIME, monitor.Time(time.Now()))

		now := time.Now()
		next := p.schedule.Next(now)
		timer := time.NewTimer(next.Sub(now))
		p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, monitor.Time(next))

		for {
			select {
			case <-p.ctx.Done():
				return
			case now = <-timer.C:
				next = p.schedule.Next(now)
				timer.Reset(next.Sub(now))
				p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, monitor.Time(next))
				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_START_TIME, monitor.Time(now))
				p.monitor.Add(METRICS_KEY_PIPELINE_RUN_TIMES, 1)

				c.Run()

				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_END_TIME, monitor.Time(time.Now()))
			}
		}
	}()

	return nil
}

func (p *pipeliner) Stop() {
	if p.isStopped() {
		return
	}

	p.cancel()

	p.runningWg.Wait()

	for _, c := range p.components {
		if err := c.Component.Stop(); err != nil {
			log.Error("Failed to stop %s component error: %s", c.Component.Instance().Name(), err)
		}
	}

	atomic.StoreInt32(&p.state, int32(Exited))
}

func (p *pipeliner) isStopped() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
	}
	return false
}

func (p *pipeliner) State() State {
	return State(atomic.LoadInt32(&p.state))
}

func (p *pipeliner) ListComponents() []Component {
	return p.components
}

func (p *pipeliner) ListProcessors() []Processor {
	return p.processors
}

func (p *pipeliner) Visualize(w io.Writer, format string) error {
	v, ok := visualizers[format]
	if !ok {
		return fmt.Errorf("Unsupported visualize type: %s, supported visualize types: %s",
			format, supportedVisualizerTypes)
	}

	return v(w, p)
}

func (p *pipeliner) Monitor() monitor.Monitor {
	return p.monitor
}

func (p *pipeliner) GetConfig() Config {
	return p.config
}

func (p *pipeliner) SetConfig(Config) error {
	return errors.New("Unimplemented method")
}
