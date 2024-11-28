package ioc

import (
	"github.com/Tuanzi-bug/tuan-book/internal/job"
	"github.com/Tuanzi-bug/tuan-book/internal/service"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(service service.RankingService, client *rlock.Client) job.Job {
	return job.NewRankingJob(service, time.Second*30, client)
}

func InitJobs(rankingJob job.Job) *cron.Cron {
	builder := job.NewCronJobBuilder(prometheus.SummaryOpts{
		Namespace: "tuan_book",
		Subsystem: "job",
		Name:      "job_duration",
		Help:      "Job duration in milliseconds",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", builder.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return expr
}
