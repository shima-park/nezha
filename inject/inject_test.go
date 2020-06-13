package inject

import (
	"fmt"

	"reflect"
	"testing"
)

type SpecialString interface {
}

type TestStruct struct {
	Dep1 string        `inject:"t" json:"-"`
	Dep2 SpecialString `inject:""`
	Dep3 string
}

type Greeter struct {
	Name string
}

func (g *Greeter) String() string {
	return "Hello, My name is" + g.Name
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type InvokeStruct struct {
	D1 string                `inject:""`
	D2 SpecialString         `inject:""`
	D3 <-chan *SpecialString `inject:""`
	D4 chan<- *SpecialString `inject:""`
}

func Test_InjectorInvoke(t *testing.T) {
	injector := New()
	expect(t, injector == nil, false)

	dep := "some dependency"
	depName := "D1"
	injector.Map(dep, depName)
	dep2 := "another dep"
	dep2Name := "D2"
	injector.MapTo(dep2, dep2Name, (*SpecialString)(nil))
	dep3 := make(chan *SpecialString)
	dep4 := make(chan *SpecialString)
	typRecv := reflect.ChanOf(reflect.RecvDir, reflect.TypeOf(dep3).Elem())
	typSend := reflect.ChanOf(reflect.SendDir, reflect.TypeOf(dep4).Elem())
	injector.Set(typRecv, "D3", reflect.ValueOf(dep3))
	injector.Set(typSend, "D4", reflect.ValueOf(dep4))

	_, err := injector.Invoke(func(i *InvokeStruct) {
		expect(t, i.D1, dep)
		expect(t, i.D2, dep2)
		expect(t, reflect.TypeOf(i.D3).Elem(), reflect.TypeOf(dep3).Elem())
		expect(t, reflect.TypeOf(i.D4).Elem(), reflect.TypeOf(dep4).Elem())
		expect(t, reflect.TypeOf(i.D3).ChanDir(), reflect.RecvDir)
		expect(t, reflect.TypeOf(i.D4).ChanDir(), reflect.SendDir)
	})

	expect(t, err, nil)
}

type InvokeStruct2 struct {
	D1 string `inject:""`
	D2 string `inject:"D1"`
}

func Test_InjectorTag(t *testing.T) {
	injector := New()
	expect(t, injector == nil, false)

	dep := "some dependency"
	depName := "D1"
	injector.Map(dep, depName)

	_, err := injector.Invoke(func(i *InvokeStruct2, i2 InvokeStruct2) {
		expect(t, i.D1, dep)
		expect(t, i.D2, dep)
		expect(t, i2.D1, dep)
		expect(t, i2.D2, dep)
	})

	expect(t, err, nil)
}

type InvokeStruct3 struct {
	D1 string        `inject:""`
	D2 SpecialString `inject:""`
}

func Test_InjectorInvokeReturnValues(t *testing.T) {
	injector := New()
	expect(t, injector == nil, false)

	dep := "some dependency"
	injector.Map(dep, "D1")
	dep2 := "another dep"
	injector.MapTo(dep2, "D2", (*SpecialString)(nil))

	result, err := injector.Invoke(func(i *InvokeStruct3) string {
		expect(t, i.D1, dep)
		expect(t, i.D2, dep2)
		return "Hello world"
	})

	expect(t, len(result), 1)
	expect(t, result[0].String(), "Hello world")
	expect(t, err, nil)
}

func Test_InjectorApply(t *testing.T) {
	injector := New()

	injector.Map("a dep", "t").MapTo("another dep", "Dep2", (*SpecialString)(nil))

	s := TestStruct{}
	err := injector.Apply(&s)
	expect(t, err, nil)

	expect(t, s.Dep1, "a dep")
	expect(t, s.Dep2, "another dep")
	expect(t, s.Dep3, "")
}

func Test_InterfaceOf(t *testing.T) {
	iType := InterfaceOf((*SpecialString)(nil))
	expect(t, iType.Kind(), reflect.Interface)

	iType = InterfaceOf((**SpecialString)(nil))
	expect(t, iType.Kind(), reflect.Interface)

	// Expecting nil
	defer func() {
		rec := recover()
		refute(t, rec, nil)
	}()
	iType = InterfaceOf((*testing.T)(nil))
	_ = iType
}

func Test_InjectorSet(t *testing.T) {
	injector := New()
	typ := reflect.TypeOf("string")
	typSend := reflect.ChanOf(reflect.SendDir, typ)
	typRecv := reflect.ChanOf(reflect.RecvDir, typ)

	// instantiating unidirectional channels is not possible using reflect
	// http://golang.org/src/pkg/reflect/value.go?s=60463:60504#L2064
	chanRecv := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)
	chanSend := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)

	injector.Set(typSend, "SendChan", chanSend)
	injector.Set(typRecv, "RecvChan", chanRecv)

	expect(t, injector.Get(typSend, "SendChan").IsValid(), true)
	expect(t, injector.Get(typRecv, "RecvChan").IsValid(), true)
	expect(t, injector.Get(chanSend.Type(), "send_chan").IsValid(), false)
}

func Test_InjectorGet(t *testing.T) {
	injector := New()

	injector.Map("some dependency", "s")

	expect(t, injector.Get(reflect.TypeOf("string"), "s").IsValid(), true)
	expect(t, injector.Get(reflect.TypeOf(11), "i").IsValid(), false)
}

func Test_InjectorSetParent(t *testing.T) {
	injector := New()
	injector.MapTo("another dep", "dep", (*SpecialString)(nil))

	injector2 := New()
	injector2.SetParent(injector)

	expect(t, injector2.Get(InterfaceOf((*SpecialString)(nil)), "dep").IsValid(), true)
}

func TestInjectImplementors(t *testing.T) {
	injector := New()
	g := &Greeter{"Jeremy"}
	injector.Map(g, "g")

	g2 := &Greeter{"Foo"}
	injector.Map(g2, "g2")

	expect(t, injector.Get(InterfaceOf((*fmt.Stringer)(nil)), "t").IsValid(), false)
	expect(t, injector.Get(InterfaceOf((*fmt.Stringer)(nil)), "g").IsValid(), true)

	expect(t, g.Name, injector.Get(InterfaceOf((*fmt.Stringer)(nil)), "g").Interface().(*Greeter).Name)
	expect(t, g2.Name, injector.Get(InterfaceOf((*fmt.Stringer)(nil)), "g2").Interface().(*Greeter).Name)
}
