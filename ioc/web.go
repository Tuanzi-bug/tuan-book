package ioc

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	"github.com/Tuanzi-bug/tuan-book/internal/web/middleware"
	"github.com/Tuanzi-bug/tuan-book/pkg/gin-plus/middlewares/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHdl.RegisterRoutes(server)
	return server
}
func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/login").Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}
func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"x-jwt-token"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			fmt.Println(origin)
			return true
		},
		//MaxAge: 12 * time.Hour,
	})
}
