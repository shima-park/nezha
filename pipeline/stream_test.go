package pipeline

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestStream(t *testing.T) {
	f := &Stream{name: "root"}

	equalsSlice(t, travel(f, 0), []string{"root"})

	if err := f.AppendByParentName("root", &Stream{name: "step1"}); err != nil {
		t.Fatal(err)
	}

	equalsSlice(t, travel(f, 0), []string{"root", "step1"})

	if err := f.InsertBefore("step1", &Stream{name: "step0"}); err != nil {
		t.Fatal(err)
	}

	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1"})

	if err := f.InsertAfter("step1", &Stream{name: "step2"}); err != nil {
		t.Fatal(err)
	}
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2"})

	if err := f.InsertAfter("step2", &Stream{name: "step3"}); err != nil {
		t.Fatal(err)
	}
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2", "step3"})

	if err := f.AppendByParentName("step1", &Stream{name: "step1.5"}); err != nil {
		t.Fatal(err)
	}
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step1.5", "step2", "step3"})

	step1, ok := f.Get("step1")
	assert.Equal(t, true, ok)
	assert.Equal(t, step1.childs[0].name, "step1.5")

	if err := f.Delete("step0"); err != nil {
		t.Fatal(err)
	}
	equalsSlice(t, travel(f, 0), []string{"root", "step1", "step1.5", "step2", "step3"})

	if err := f.Delete("root"); err != nil {
		t.Fatal(err)
	}
	equalsSlice(t, travel(f, 0), []string{"root"})
}

func equalsSlice(t *testing.T, actual, expected []string) {
	assert.Equal(t, strings.Join(actual, ","), strings.Join(expected, ","))
}
