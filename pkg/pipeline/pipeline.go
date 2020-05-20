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
	components []NamedComponent
	injector   inject.Injector
	stream     *Stream

	state     int32
	ctrlWg    sync.WaitGroup
	runningWg sync.WaitGroup
	done      chan struct{}
}

type NamedComponent struct {
	Name      string
	RawConfig string
	Component component.Component
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

	for _, c := range p.components {
		instance := c.Component.Instance()

		p.injector.Set(instance.Type(), instance.Name(), instance.Value())
	}

	return p, nil
}

func NewPipelineByConfig(conf Config, opts ...Option) (*Pipeline, error) {
	var components []NamedComponent
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
			components = append(components, NamedComponent{
				Name:      componentName,
				RawConfig: rawConfig,
				Component: c,
			})
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

func (p *Pipeline) CheckDependence() []error {
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
	if !atomic.CompareAndSwapInt32(&p.state, int32(Idle), int32(Running)) {
		return nil
	}

	for _, c := range p.components {
		if err := c.Component.Start(); err != nil {
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
	if atomic.CompareAndSwapInt32(&p.state, int32(Running), int32(Waiting)) {
		p.ctrlWg.Add(1)
	}
}

func (p *Pipeline) Resume() {
	if atomic.CompareAndSwapInt32(&p.state, int32(Waiting), int32(Running)) {
		p.ctrlWg.Done()
	}
}

func (p *Pipeline) Stop() {
	if p.isStopped() {
		return
	}

	close(p.done)
	atomic.StoreInt32(&p.state, int32(Closed))
	p.runningWg.Wait()

	for _, c := range p.components {
		if err := c.Component.Stop(); err != nil {
			log.Error("Failed to stop %s component error: %s", c.Component.Instance().Name(), err)
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

func (p *Pipeline) State() string {
	return state(p.state).String()
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

type ComponentView struct {
	Name         string
	RawConfig    string
	SampleConfig string
	Description  string
	InjectName   string
	ReflectType  string
	ReflectValue string
}

func (p *Pipeline) ListComponent() []ComponentView {
	var list []ComponentView
	for _, c := range p.components {
		i := c.Component.Instance()
		list = append(list, ComponentView{
			Name:         c.Name,
			RawConfig:    c.RawConfig,
			SampleConfig: c.Component.SampleConfig(),
			Description:  c.Component.Description(),
			InjectName:   i.Name(),
			ReflectType:  i.Type().String(),
			ReflectValue: i.Value().String(),
		})
	}
	return list
}

type ProcessorView struct {
	Name            string
	ProcessorName   string
	ProcessorConfig string
	Request         []Receptor
	Response        []Receptor
}

func (p *Pipeline) ListProcessor() []ProcessorView {
	var list []ProcessorView
	listProcessor(p.stream, &list)
	return list
}

func listProcessor(stream *Stream, list *[]ProcessorView) {
	if stream == nil {
		return
	}

	req, resp := getFuncReqAndRespReceptorList(stream.processor)

	(*list) = append((*list), ProcessorView{
		Name:            stream.name,
		ProcessorName:   stream.processorName,
		ProcessorConfig: stream.processorConfig,
		Request:         req,
		Response:        resp,
	})

	for i := 0; i < len(stream.childs); i++ {
		listProcessor(stream.childs[i], list)
	}
}
