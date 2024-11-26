package main

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/ioc"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"net/http"
)

func main() {
	// 使用viper配置管理
	initViper()
	// 启动日志
	ioc.InitLogger()
	app := InitWebServer()
	initPrometheus()
	server := app.server
	//启动定时任务
	log.Info("start cron jobs")
	app.cron.Start()
	defer func() {
		// 等待定时任务结束
		<-app.cron.Stop().Done()
	}()
	// 启动消费者
	log.Info("start consumers")
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	// 启动web服务
	log.Info("start web server")
	_ = server.Run(":9080")
}

func initViper() {
	viper.SetConfigFile("./config/dev.yaml") // 指定配置文件路径
	err := viper.ReadInConfig()              // 读取配置信息
	if err != nil {                          // 读取配置信息失败
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// 监控配置文件变化
	viper.WatchConfig()
}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9081", nil)
	}()
}
