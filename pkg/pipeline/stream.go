package pipeline

import (
	"errors"
	"nezha/pkg/common/log"
	"nezha/pkg/processor"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Stream struct {
	rwlock    sync.RWMutex
	name      string
	processor processor.Processor
	parent    *Stream
	childs    []*Stream
}

func NewStream(conf StreamConfig) (*Stream, error) {
	p, err := processor.New(conf.ProcessorName, conf.ProcessorConfig)
	if err != nil {
		return nil, err
	}

	f := &Stream{
		name:      conf.Name,
		processor: p,
	}

	if f.name == "" {
		f.name = uuid.New().String()
	}

	for _, subConf := range conf.Childs {
		subStream, err := NewStream(subConf)
		if err != nil {
			return nil, err
		}
		f.Append(subStream)
	}

	return f, nil
}

func (f *Stream) Name() string {
	return f.name
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
	_, _, ok := get(f, s.name, 0)
	if ok {
		return errors.New("The " + s.name + " stream is exists")
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
	_, _, ok := get(f, s.name, 0)
	if ok {
		return errors.New("The " + s.name + " stream is exists")
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
	_, _, ok := get(f, s.name, 0)
	if ok {
		return errors.New("The " + s.name + " stream is exists")
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
		if c.name == name {
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

func get(f *Stream, name string, i int) (*Stream, int, bool) {
	if f == nil {
		return nil, 0, false
	}

	if f.name == name {
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

	var arr = []string{f.name}
	log.Info("%s%s", strings.Repeat(" ", depth), f.name)

	depth += 4
	for i := 0; i < len(f.childs); i++ {
		c := f.childs[i]
		res := travel(c, depth)
		arr = append(arr, res...)
	}

	return arr
}
