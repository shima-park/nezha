package processor

import (
	"errors"
	"reflect"
)

type Processor interface{}

func Validate(processor Processor) error {
	if reflect.TypeOf(processor).Kind() != reflect.Func {
		return errors.New("Processor must be a callable func")
	}
	return nil
}
