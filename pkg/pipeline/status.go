package pipeline

import "fmt"

var (
	Idle    status = 0
	Running status = 1
	Waiting status = 2
	Closed  status = 3
)

type status int32

func (s status) String() string {
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
