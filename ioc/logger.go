package ioc

import (
	"github.com/Tuanzi-bug/tuan-book/pkg/log"
)

func InitLogger() {
	// logger配置
	opts := &log.Options{
		Development:   true,
		Level:         "debug",
		Format:        "console",
		EnableColor:   true, // if you need output to local path, with EnableColor must be false.
		DisableCaller: true,
		OutputPaths:   []string{"stdout"},
		//ErrorOutputPaths: []string{"error.log"},
	}
	// 初始化全局logger
	log.Init(opts)
}
