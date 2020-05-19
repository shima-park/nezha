package pipeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"

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

	status    int32
	ctrlWg    sync.WaitGroup
	runningWg sync.WaitGroup
	done      chan struct{}
}

func New(opts ...Option) (*Pipeline, error) {
	p := &Pipeline{
		ctx:      context.Background(),
		id:       uuid.New().String(),
		injector: inject.New(),
		done:     make(chan struct{}),
	}

	apply(p, opts)

	p.injector = NewLoggerInjector("Pipeline: "+p.Name(), p.injector)
	p.injector.MapTo(p.ctx, "Context", (*context.Context)(nil))

	if p.stream == nil {
		return nil, errors.Wrap(errors.New("The pipeline must have at least one stream"), p.Name())
	}

	for _, component := range p.components {
		instance := component.Instance()

		p.injector.Set(instance.Type(), instance.Name(), instance.Value())
	}

	if err := p.checkDependence(); err != nil {
		return nil, errors.Wrapf(err, "Pipeline(%s)", p.Name())
	}

	return p, nil
}

func NewPipelineByConfig(conf Config, opts ...Option) (*Pipeline, error) {
	var components []component.Component
	for _, name2config := range conf.Components {
		for componentName, rawConfig := range name2config {
			factory, err := component.GetFactory(componentName)
			if err != nil {
				return nil, err
			}
			c, err := factory(rawConfig)
			if err != nil {
				return nil, err
			}
			components = append(components, c)
		}
	}

	stream, err := NewStream(*conf.Stream)
	if err != nil {
		return nil, err
	}

	return New(
		append(
			[]Option{
				WithName(conf.Name),
				WithComponents(components...),
				WithStream(stream),
			},
			opts...,
		)...,
	)
}

func (p *Pipeline) checkDependence() error {
	checkInj := inject.New()
	checkInj.SetParent(p.injector)
	checkInj = NewLoggerInjector(fmt.Sprintf("Pipeline(%s)", p.Name()), checkInj)
	return check(p.stream, checkInj)
}

func (p *Pipeline) ID() string {
	return p.id
}

func (p *Pipeline) Name() string {
	return p.name
}

func (p *Pipeline) Start() error {
	if !atomic.CompareAndSwapInt32(&p.status, int32(Idle), int32(Running)) {
		return nil
	}

	for _, c := range p.components {
		if err := c.Start(); err != nil {
			return err
		}
	}

	p.runningWg.Add(1)
	defer p.runningWg.Done()

	for !p.isStopped() {
		select {
		case <-p.ctx.Done():
			p.Stop()
		default:
		}

		p.ctrlWg.Wait()

		ctx, cancel := context.WithCancel(p.ctx)
		defer cancel()

		c := NewExecContext(ctx, p.stream, p.injector)
		if err := c.Run(); err != nil {
			log.Error("[Pipeline] error: %v", err)
		}
	}

	return nil
}

func (p *Pipeline) Wait() {
	if atomic.CompareAndSwapInt32(&p.status, int32(Running), int32(Waiting)) {
		p.ctrlWg.Add(1)
	}
}

func (p *Pipeline) Resume() {
	if atomic.CompareAndSwapInt32(&p.status, int32(Waiting), int32(Running)) {
		p.ctrlWg.Done()
	}
}

func (p *Pipeline) Stop() {
	if p.isStopped() {
		return
	}

	close(p.done)
	atomic.StoreInt32(&p.status, int32(Closed))
	p.runningWg.Wait()

	for _, c := range p.components {
		if err := c.Stop(); err != nil {
			log.Error("Failed to stop %s component error: %s", c.Instance().Name(), err)
		}
	}
}

func (p *Pipeline) isStopped() bool {
	select {
	case <-p.done:
		return true
	default:
	}
	return false
}

func (p *Pipeline) Status() string {
	return status(p.status).String()
}

// 单词执行并尝试从容器中捞回一些执行过程中的数据
func (p *Pipeline) Exec(ctx context.Context, ret interface{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := NewExecContext(ctx, p.stream, p.injector)
	if err := c.Run(); err != nil {
		log.Error("[Pipeline] error: %v", err)
	}

	return c.injector.Apply(ret)
}

func (p *Pipeline) ListComponent() []string {
	var list []string
	for _, component := range p.components {
		i := component.Instance()
		list = append(list, fmt.Sprintf("Name: %s, Type: %s", i.Name(), i.Type()))
	}
	return list
}
