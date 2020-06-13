package gin

import "testing"

func TestGin(t *testing.T) {
	g, err := NewGin(`
      name: "GinServer"
      addr: ":8080"
      routers:
        GET: /send/article`)
	if err != nil {
		panic(err)
	}
	g.Start()
	select {}
}
