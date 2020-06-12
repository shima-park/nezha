package config

import "gopkg.in/yaml.v2"

var (
	DefaultUnmarshal UnmarshalFunc = yaml.Unmarshal
	DefaultMarshal   MarshalFunc   = yaml.Marshal
)

type UnmarshalFunc func(in []byte, v interface{}) error

type MarshalFunc func(in interface{}) ([]byte, error)

func Unmarshal(in []byte, v interface{}) error {
	return DefaultUnmarshal(in, v)
}

func Marshal(in interface{}) ([]byte, error) {
	return DefaultMarshal(in)
}

func MustMarshal(v interface{}) string {
	b, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
