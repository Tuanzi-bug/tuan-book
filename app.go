package main

import (
	"github.com/Tuanzi-bug/tuan-book/internal/events"
	"github.com/gin-gonic/gin"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}
