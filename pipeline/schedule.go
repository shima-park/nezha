package pipeline

import (
	"time"

	"github.com/robfig/cron/v3"
)

var standardParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

var defaultSchedule cron.Schedule = ConstantDelaySchedule{}

type ConstantDelaySchedule struct {
}

func (schedule ConstantDelaySchedule) Next(t time.Time) time.Time {
	return t
}
