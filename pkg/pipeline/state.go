package pipeline

import "fmt"

var (
	Idle    state = 0
	Running state = 1
	Waiting state = 2
	Exited  state = 3
)

type state int32

func (s state) String() string {
	switch s {
	case Idle:
		return "idle"
	case Running:
		return "running"
	case Waiting:
		return "waiting"
	case Exited:
		return "exited"
	}
	return fmt.Sprintf("unknown(%d)", s)
}
