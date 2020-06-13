package pipeline

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestStream(t *testing.T) {
	f := &Stream{processor: Processor{Name: "root"}}

	equalsSlice(t, travel(f, 0), []string{"root"})

	err := f.AppendByParentName("root", &Stream{processor: Processor{Name: "step1"}})
	assert.NilError(t, err)

	equalsSlice(t, travel(f, 0), []string{"root", "step1"})

	err = f.InsertBefore("step1", &Stream{processor: Processor{Name: "step0"}})
	assert.NilError(t, err)

	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1"})

	err = f.InsertAfter("step1", &Stream{processor: Processor{Name: "step2"}})
	assert.NilError(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2"})

	err = f.InsertAfter("step2", &Stream{processor: Processor{Name: "step3"}})
	assert.NilError(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step2", "step3"})

	err = f.AppendByParentName("step1", &Stream{processor: Processor{Name: "step1.5"}})
	assert.NilError(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step0", "step1", "step1.5", "step2", "step3"})

	step1, ok := f.Get("step1")
	assert.Equal(t, true, ok)
	assert.Equal(t, step1.childs[0].Name(), "step1.5")

	err = f.Delete("step0")
	assert.NilError(t, err)
	equalsSlice(t, travel(f, 0), []string{"root", "step1", "step1.5", "step2", "step3"})

	err = f.Delete("root")
	assert.NilError(t, err)
	equalsSlice(t, travel(f, 0), []string{"root"})
}

func equalsSlice(t *testing.T, actual, expected []string) {
	assert.Equal(t, strings.Join(actual, ","), strings.Join(expected, ","))
}
