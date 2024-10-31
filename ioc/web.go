package ioc

import (
	"fmt"
	"github.com/Tuanzi-bug/tuan-book/internal/web"
	myjwt "github.com/Tuanzi-bug/tuan-book/internal/web/jwt"
	"github.com/Tuanzi-bug/tuan-book/internal/web/middleware"
	"github.com/Tuanzi-bug/tuan-book/pkg/gin-plus/middlewares/ratelimit"
	"github.com/Tuanzi-bug/tuan-book/pkg/limiter"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHdl *web.UserHandler, artHandler *web.ArticleHandler) *gin.Engine {
	// 因为重写了log和recovery中间件
	server := gin.New()
	server.Use(middlewares...)
	userHdl.RegisterRoutes(server)
	artHandler.RegisterRoutes(server)
	return server
}
func InitMiddlewares(redisClient redis.Cmdable, hdl myjwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		//// 加入zap中间件
		//middleware.Ginzap(logger, time.RFC3339, true),
		//// 加入Recovery中间件
		//middleware.RecoveryWithZap(logger, true),
		corsHdl(),
		middleware.NewJWTMiddlewareBuilder(hdl).
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/login").
			IgnorePaths("/users/refresh_token").Build(),
		// 采用滑动窗口算法构建限流器：1s内允许1000个请求。具体的数值需要根据压测来决定
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
	}
}
func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			fmt.Println(origin)
			return true
		},
		//MaxAge: 12 * time.Hour,
	})
}
