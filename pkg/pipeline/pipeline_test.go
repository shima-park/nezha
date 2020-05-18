package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	p, err := New()
	assert.NotNil(t, err)
	assert.Nil(t, p)

	p2, err := New(WithStream(&Stream{}))
	assert.Nil(t, err)
	assert.NotNil(t, p2)
}
