package pipeline

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/shima-park/nezha/common/log"
	"github.com/shima-park/nezha/inject"
	"github.com/shima-park/nezha/processor"
)

type Stream struct {
	rwlock    sync.RWMutex
	processor Processor
	parent    *Stream
	childs    []*Stream
	config    StreamConfig
}

func NewStream(conf StreamConfig, processors map[string]Processor) (*Stream, error) {
	p, ok := processors[conf.Name]
	if !ok {
		return nil, fmt.Errorf("Not found processor %s", conf.Name)
	}

	if conf.Replica == 0 {
		conf.Replica = 1
	}

	f := &Stream{
		processor: p,
		config:    conf,
	}

	for _, subConf := range conf.Childs {
		subStream, err := NewStream(subConf, processors)
		if err != nil {
			return nil, err
		}
		f.Append(subStream)
	}

	return f, nil
}

func (f *Stream) Name() string {
	return f.processor.Name
}

func (f *Stream) Append(s *Stream) *Stream {
	f.rwlock.Lock()
	defer f.rwlock.Unlock()
	s.parent = f
	f.childs = append(f.childs, s)
	return f
}

func (f *Stream) AppendByParentName(parentName string, s *Stream) error {
	f.rwlock.Lock()
	defer f.rwlock.Unlock()
	_, _, ok := get(f, s.Name(), 0)
	if ok {
		return errors.New("The " + s.Name() + " stream is exists")
	}

	p, _, ok := get(f, parentName, 0)
	if !ok {
		return errors.New("Can't find stream's parent " + parentName)
	}

	s.parent = p
	p.childs = append(p.childs, s)
	return nil
}

func (f *Stream) InsertBefore(broName string, s *Stream) error {
	f.rwlock.Lock()
	defer f.rwlock.Unlock()
	_, _, ok := get(f, s.Name(), 0)
	if ok {
		return errors.New("The " + s.Name() + " stream is exists")
	}

	bro, index, ok := get(f, broName, 0)
	if !ok {
		return errors.New("Can't find stream's brother " + broName)
	}

	p := bro.parent
	s.parent = p
	p.childs = append(p.childs, s)
	swim(p.childs, index, len(p.childs)-1)
	return nil
}

// 上浮 将k元素上浮至j元素的位置
func swim(streams []*Stream, j, k int) {
	for ; j < k; j++ {
		streams[j], streams[k] = streams[k], streams[j]
	}
}

func (f *Stream) InsertAfter(broName string, s *Stream) error {
	f.rwlock.Lock()
	defer f.rwlock.Unlock()
	_, _, ok := get(f, s.Name(), 0)
	if ok {
		return errors.New("The " + s.Name() + " stream is exists")
	}

	bro, index, ok := get(f, broName, 0)
	if !ok {
		return errors.New("Can't find stream's brother " + broName)
	}

	p := bro.parent
	s.parent = p
	p.childs = append(p.childs, s)
	swim(p.childs, index+1, len(p.childs)-1)
	return nil
}

func (f *Stream) Delete(name string) error {
	f.rwlock.Lock()
	defer f.rwlock.Unlock()
	t, i, ok := get(f, name, 0)
	if !ok {
		return errors.New("The " + name + " stream is not exists")
	}

	if t.parent == nil {
		t.childs = nil
		return nil
	}

	t.parent.childs = append(t.parent.childs[:i], t.parent.childs[i+1:]...)

	return nil
}

func (f *Stream) findChild(name string) (*Stream, int, bool) {
	for i := 0; i < len(f.childs); i++ {
		c := f.childs[i]
		if c.Name() == name {
			return c, i, true
		}
	}
	return nil, 0, false
}

func (f *Stream) Get(name string) (*Stream, bool) {
	f.rwlock.RLock()
	defer f.rwlock.RUnlock()
	f, _, ok := get(f, name, 0)
	return f, ok
}

func (f *Stream) Invoke(inj inject.Injector) (outVal reflect.Value, err error) {
	defer f.Recover(nil)

	if f == nil || f.processor.Processor == nil {
		return
	}

	p := f.processor.Processor

	err = processor.Validate(p)
	if err != nil {
		err = errors.Wrapf(err, "Stream(%s)", f.Name())
		return
	}

	var vals []reflect.Value
	vals, err = inj.Invoke(p)
	if err != nil {
		err = errors.Wrapf(err, "Stream(%s)", f.Name())
		return
	}

	return tryGetValueAndError(vals)
}

func (s *Stream) Recover(f func()) {
	if r := recover(); r != nil {
		log.Error("Stream: %s, Panic: %s, Stack: %s",
			s.Name(), r, string(debug.Stack()))
	}

	if f != nil {
		f()
	}
}

func tryGetValueAndError(vals []reflect.Value) (outVal reflect.Value, err error) {
	if len(vals) == 1 {
		// 判断一个返回值时是否时error
		if vals[0].Type().Implements(errorInterface) && !vals[0].IsNil() {
			err = vals[0].Interface().(error)
			return
		}
		// 不是error作为return value处理
		outVal = vals[0]
		return
	}

	if len(vals) == 2 {
		// 返回值为两个时候，默认认为第一个为return value
		outVal = vals[0]
		// 第二个为error
		if vals[1].Type().Implements(errorInterface) && !vals[1].IsNil() {
			err = vals[1].Interface().(error)
			return
		}
		return
	}

	return
}

func get(f *Stream, name string, i int) (*Stream, int, bool) {
	if f == nil {
		return nil, 0, false
	}

	if f.Name() == name {
		return f, i, true
	}

	for i := 0; i < len(f.childs); i++ {
		c := f.childs[i]
		target, index, ok := get(c, name, i)
		if ok {
			return target, index, true
		}
	}
	return nil, 0, false
}

func travel(f *Stream, depth int) []string {
	if f == nil {
		return nil
	}

	var arr = []string{f.Name()}
	log.Info("%s%s", strings.Repeat(" ", depth), f.Name())

	depth += 4
	for i := 0; i < len(f.childs); i++ {
		c := f.childs[i]
		res := travel(c, depth)
		arr = append(arr, res...)
	}

	return arr
}
