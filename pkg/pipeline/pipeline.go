package pipeline

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"
	"github.com/shima-park/nezha/pkg/monitor"
	"github.com/shima-park/nezha/pkg/processor"

	"github.com/shima-park/nezha/pkg/inject"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Pipeline struct {
	ctx    context.Context
	cancel context.CancelFunc

	id         string
	name       string
	components []NamedComponent
	processors []NamedProcessor
	injector   inject.Injector
	stream     *Stream
	monitor    monitor.Monitor
	startTime  time.Time
	schedule   cron.Schedule

	state     int32
	runningWg sync.WaitGroup
}

type NamedComponent struct {
	Name      string
	RawConfig string
	Component component.Component
}

type NamedProcessor struct {
	Name      string
	RawConfig string
	Processor processor.Processor
}

func New(opts ...Option) (*Pipeline, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pipeline{
		ctx:       ctx,
		cancel:    cancel,
		id:        uuid.New().String(),
		injector:  inject.New(),
		startTime: time.Now(),
	}

	apply(p, opts)

	p.monitor = monitor.NewMonitor(p.Name())
	p.monitor.Set(METRICS_KEY_PIPELINE_STATE, expvar.Func(func() interface{} { return p.State() }))

	p.injector.MapTo(p.monitor, "Monitor", (*monitor.Monitor)(nil))
	p.injector.MapTo(p.ctx, "Context", (*context.Context)(nil))

	if p.stream == nil {
		return nil, errors.Wrap(errors.New("The pipeline must have at least one stream"), p.Name())
	}

	for _, c := range p.components {
		instance := c.Component.Instance()

		p.injector.Set(instance.Type(), instance.Name(), instance.Value())
	}

	if err := p.SetRouter(); err != nil {
		return nil, err
	}

	return p, nil
}

type Request struct {
	Gin *gin.Engine `inject:"gin_server""`
}

func (p *Pipeline) SetRouter() error {
	_, err := p.injector.Invoke(func(r *Request) {
		g := r.Gin
		g.GET("/debug/visualize", func(ctx *gin.Context) {
			dotFile, err := ioutil.TempFile("", "dot")
			if err != nil {
				return
			}
			defer os.Remove(dotFile.Name())
			defer dotFile.Close()

			p.Visualize(dotFile)

			svgFile, err := ioutil.TempFile("", "svg")
			if err != nil {
				return
			}
			defer os.Remove(svgFile.Name())
			defer svgFile.Close()

			err = exec.Command("dot", "-Tpng", dotFile.Name(), "-o", svgFile.Name()).Run()
			if err != nil {
				return
			}

			b, _ := ioutil.ReadAll(svgFile)

			ctx.Writer.Write(b)
		})
		g.GET("/debug/components", func(ctx *gin.Context) {
			PrintPipelineComponents(ctx.Writer, p)
		})
		g.GET("/debug/processors", func(ctx *gin.Context) {
			PrintPipelineProcessor(ctx.Writer, p)
		})
	})

	return err
}

func NewPipelineByConfig(conf Config, opts ...Option) (*Pipeline, error) {
	components, err := conf.NewComponents()
	if err != nil {
		return nil, err
	}

	processors, err := conf.NewProcessors()
	if err != nil {
		return nil, err
	}

	var pm = map[string]NamedProcessor{}
	for _, p := range processors {
		pm[p.Name] = p
	}

	stream, err := NewStream(*conf.Pipeline.Stream, pm)
	if err != nil {
		return nil, err
	}

	var schedule = defaultSchedule
	if conf.Pipeline.Schedule != "" {
		schedule, err = standardParser.Parse(conf.Pipeline.Schedule)
		if err != nil {
			return nil, err
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
			},
			opts...,
		)...,
	)
}

func (p *Pipeline) CheckDependence() []error {
	checkInj := inject.New()
	checkInj.SetParent(p.injector)
	return check(p.stream, checkInj)
}

func (p *Pipeline) ID() string {
	return p.id
}

func (p *Pipeline) Name() string {
	return p.name
}

func (p *Pipeline) newExecContext() *execContext {
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

func (p *Pipeline) Start() error {
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
			p.monitor.Set(METRICS_KEY_PIPELINE_UPTIME, Elapsed(time.Since(p.startTime)))
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

			p.monitor.Set(METRICS_KEY_PIPELINE_EXIT_TIME, Time(time.Now()))

			p.runningWg.Done()
		}()

		p.monitor.Set(METRICS_KEY_PIPELINE_START_TIME, Time(time.Now()))

		now := time.Now()
		next := p.schedule.Next(now)
		timer := time.NewTimer(next.Sub(now))
		p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, Time(next))

		for {
			select {
			case <-p.ctx.Done():
				return
			case now = <-timer.C:
				next = p.schedule.Next(now)
				timer.Reset(next.Sub(now))
				p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, Time(next))
				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_START_TIME, Time(now))
				p.monitor.Add(METRICS_KEY_PIPELINE_RUN_TIMES, 1)

				c.Run()

				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_END_TIME, Time(time.Now()))
			}
		}
	}()

	return nil
}

func (p *Pipeline) Stop() {
	if p.isStopped() {
		return
	}

	p.cancel()

	p.runningWg.Wait()

	atomic.StoreInt32(&p.state, int32(Exited))

	for _, c := range p.components {
		if err := c.Component.Stop(); err != nil {
			log.Error("Failed to stop %s component error: %s", c.Component.Instance().Name(), err)
		}
	}
}

func (p *Pipeline) isStopped() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
	}
	return false
}

func (p *Pipeline) State() string {
	return state(atomic.LoadInt32(&p.state)).String()
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
	Processor NamedProcessor
	Request   []Receptor
	Response  []Receptor
}

func (p *Pipeline) ListProcessor() []ProcessorView {
	var list []ProcessorView

	for _, p := range p.processors {
		req, resp := getFuncReqAndRespReceptorList(p.Processor)

		list = append(list, ProcessorView{
			Processor: p,
			Request:   req,
			Response:  resp,
		})
	}
	return list
}

func (p *Pipeline) Visualize(w io.Writer) error {
	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	buffer.WriteString(`node [shape=plaintext fontname="Sans serif" fontsize="24"];` + "\n")

	buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
		p.Name(),
	))

	first := true
	p.monitor.Do(func(namespace string, kv expvar.KeyValue) {
		if namespace != p.Name() {
			return
		}
		if first {
			first = false
			buffer.WriteString("<tr><td align=\"left\"><b>" + p.Name() + "</b></td></tr>\n")
		}

		buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
	})
	buffer.WriteString("</table>>];\n")
	buffer.WriteString("\n")

	for _, proc := range p.processors {
		buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
			proc.Name,
		))

		first := true
		p.monitor.Do(func(namespace string, kv expvar.KeyValue) {
			if namespace != proc.Name {
				return
			}
			if first {
				first = false
				buffer.WriteString("<tr><td align=\"left\"><b>" + proc.Name + "</b></td></tr>\n")
			}

			buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
		})

		buffer.WriteString("</table>>];\n")
		buffer.WriteString("\n")
	}

	p.visualize(p.stream, &buffer)

	buffer.WriteString("}")
	_, err := w.Write(buffer.Bytes())
	return err
}

func (p *Pipeline) visualize(s *Stream, w io.Writer) {
	if s == nil {
		return
	}

	for _, x := range s.childs {
		w.Write([]byte(fmt.Sprintf("  %s %s %s;\n", s.Name(), "->", x.Name())))
		p.visualize(x, w)
	}
}

func (p *Pipeline) Monitor() monitor.Monitor {
	return p.monitor
}
