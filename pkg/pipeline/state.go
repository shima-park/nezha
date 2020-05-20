package pipeline

import "fmt"

var (
	Idle    state = 0
	Running state = 1
	Waiting state = 2
	Closed  state = 3
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
	case Closed:
		return "closed"
	}
	return fmt.Sprintf("unknown(%d)", s)
}
