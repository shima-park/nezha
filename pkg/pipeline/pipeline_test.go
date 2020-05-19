package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	p, err := New()
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p2, err := New(WithStream(&Stream{
		name: "test_stream",
		processor: func() {
			time.Sleep(time.Second)
		},
	}))
	assert.Nil(t, err)
	assert.NotNil(t, p2)
	assert.Equal(t, int32(Idle), p2.status)

	go func() {
		err = p2.Start()
		assert.Nil(t, err)
	}()
	time.Sleep(time.Second)
	assert.Equal(t, int32(Running), p2.status)

	p2.Wait()
	assert.Equal(t, int32(Waiting), p2.status)

	p2.Resume()
	assert.Equal(t, int32(Running), p2.status)

	p2.Stop()
	assert.Equal(t, int32(Closed), p2.status)
}
