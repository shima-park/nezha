package config

import "encoding/json"

func Unmarshal(in []byte, v interface{}) error {
	return json.Unmarshal(in, v)
}

func Marshal(in interface{}) ([]byte, error) {
	return json.Marshal(in)
}
