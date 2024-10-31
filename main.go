package main

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/ioc"
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	// 使用viper配置管理
	initViper()
	// 启动日志
	ioc.InitLogger()
	app := InitWebServer()
	server := app.server
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
	_ = server.Run(":8080")
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
