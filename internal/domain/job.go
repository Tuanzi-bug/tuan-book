package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id   int64
	Name string

	// The number of times the job has been executed
	Expression string
	Executor   string
	Cfg        string
	CancelFunc func()
}

// NextTime returns the next time the job will be executed
func (j Job) NextTime() time.Time {
	c := cron.NewParser(cron.Second | cron.Minute | cron.Hour |
		cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s, _ := c.Parse(j.Expression)
	return s.Next(time.Now())
}
