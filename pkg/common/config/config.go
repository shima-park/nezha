package config

import "encoding/json"

var (
	DefaultUnmarshal UnmarshalFunc = json.Unmarshal
	DefaultMarshal   MarshalFunc   = json.Marshal
)

type UnmarshalFunc func(in []byte, v interface{}) error

type MarshalFunc func(in interface{}) ([]byte, error)

func Unmarshal(in []byte, v interface{}) error {
	return DefaultUnmarshal(in, v)
}

func Marshal(in interface{}) ([]byte, error) {
	return DefaultMarshal(in)
}
