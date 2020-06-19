package client

import "github.com/shima-park/nezha/pkg/rpc/proto"

type server struct {
	apiBuilder
}

func (s *server) Metadata() (proto.MetadataView, error) {
	var ret proto.MetadataView
	err := GetJSON(s.api("/metadata"), &ret)
	return ret, err
}
