package ioc

import (
	"github.com/Tuanzi-bug/tuan-book/internal/web/middleware"
	"go.uber.org/zap"
)

func InitLogger() middleware.ZapLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}
