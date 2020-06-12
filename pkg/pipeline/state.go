package pipeline

import "fmt"

var (
	Idle    State = 0
	Running State = 1
	Waiting State = 2
	Exited  State = 3
)

type State int32

func (s State) String() string {
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
