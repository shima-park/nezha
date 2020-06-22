package main

import (
	"bytes"
	"fmt"

	"github.com/shima-park/nezha/pkg/util/editor"
	"gopkg.in/yaml.v2"
)

type Foo struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func main() {
	edit := editor.NewDefaultEditor(nil)

	b, err := yaml.Marshal(Foo{Name: "foo", Age: 18})
	if err != nil {
		panic(err)
	}

	buff := bytes.NewBuffer(b)

	editedDiff := b
	edited, file, err := edit.LaunchTempFile("test-edit-", "", buff)
	if err != nil {
		panic(err)
	}

	if bytes.Equal(edited, editedDiff) {
		fmt.Println("Edit cancelled, no changes made.")
		return
	}

	fmt.Println(file)
	fmt.Println(string(edited))
}
