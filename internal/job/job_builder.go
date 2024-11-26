package job

import (
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt, []string{"job", "success"})
	return &CronJobBuilder{vector: vector}
}

func (c *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		start := time.Now()
		log.Debug("开始运行任务", log.String("name", name), log.Time("start", start))
		err := job.Run()
		if err != nil {
			log.Error("任务运行失败", log.String("name", name), log.Err(err))
		}
		log.Debug("任务运行结束", log.String("name", name), log.Duration("cost", time.Since(start)))
		duration := time.Since(start).Milliseconds()
		c.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration))
	})
}

type cronJobAdapterFunc func()

func (c cronJobAdapterFunc) Run() {
	c()
}
