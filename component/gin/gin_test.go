package gin

import (
	"testing"

	"gotest.tools/assert"
)

func TestGin(t *testing.T) {
	g, err := NewGin(`
      name: "GinServer"
      addr: ":8080"
      routers:
        GET: /send/article`)
	if err != nil {
		panic(err)
	}
	err = g.Start()
	assert.NilError(t, err)
	err = g.Stop()
	assert.NilError(t, err)
}
